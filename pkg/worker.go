package pkg

import (
	"context"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	. "github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"log"
	"sync"
)

type IWorker[Data DataEntity] interface {
	Init()
	Run()
	Handle(msg Message) error
}

type Worker[Data DataEntity] struct {
	internal.Base[Data]
	ReInit        func()
	ReRun         func()
	ReHandle      func(msg Message) error
	MaxGoroutines int
}

func (w *Worker[Data]) Init() {
	if w.ReInit != nil {
		w.ReInit()
		return
	}

	if err := w.InitDB((*w.Config)[w.GetDBSection()].(util.Config)); err != nil {
		panic(err)
	}

	if err := w.InitMQ((*w.Config)[w.GetMQSection()].(util.Config)); err != nil {
		panic(err)
	}

	w.IMQ.Start() // Start Consumer
}

func (w *Worker[Data]) Run() {
	if w.ReRun != nil {
		w.ReRun()
		return
	}

	go func() {
		var wg sync.WaitGroup
		sem := make(chan struct{}, w.MaxGoroutines)
		for {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer func() { wg.Done(); <-sem }()

				msg := w.FetchMessage(context.Background())
				if w.DEBUG {
					log.Printf("[FSM] fetch msg %s", msg.Body)
				}
				if err := w.Handle(msg); err != nil {
					log.Printf("[FSM] handle %s Err: %v", msg.Body, err)
				}
			}()
		}
	}()
}

func (w *Worker[Data]) Handle(msg Message) (err error) {
	if w.ReHandle != nil {
		return w.ReHandle(msg)
	}

	defer func() {
		if err != nil {
			if msg.Nack != nil {
				if e := msg.Nack(); e != nil {
					log.Printf("[FSM] NACK %s Err: %v", msg.Body, e)
				}
			}
			return
		}
		if msg.Ack != nil {
			if e := msg.Ack(); e != nil {
				log.Printf("[FSM] ACK %s Err: %v", msg.Body, e)
			}
		}
	}()

	c := msg.C
	taskID := msg.Body

	state, err := internal.QueryTaskState(c, w.GetDB(), w.Models, taskID)
	if err != nil {
		log.Printf("[FSM] query task %s %s Err: %v", taskID, *state, err)
		return err
	}

	handler, exist := w.FSM.GetState(*state)
	if !exist {
		return nil
	}
	if handler.IsFinalState() {
		return nil
	}

	data, _ := util.Assert[Data](util.ReflectNew(w.DataModel))
	task := GenTaskInstance("", taskID, data)
	task.WithDB = w.GetDB()

	if err = internal.QueryTask(c, w.Models, task); err != nil {
		return err
	}

	if w.DEBUG {
		log.Printf("[FSM] load task %s %s %s", task.ID, task.State, util.Pretty(task))
	}
	if err = handler.Handle(task); err != nil {
		return err
	}

	task.RequestID = w.GenID()
	if err = internal.UpdateTask(c, w.Models, task, w.FSM); err != nil {
		return err
	}

	if err = w.PublishMessage(c, taskID); err != nil {
		return err
	}

	if w.DEBUG {
		log.Printf("[FSM] finish task %s %s -> %s %v", task.ID, *state, task.State, util.Pretty(task))
	}
	return nil
}
