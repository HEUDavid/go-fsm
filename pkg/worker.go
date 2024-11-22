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
				if err := w.Handle(msg); err != nil {
					log.Printf("[FSM] Handle %s ERROR: %v", msg.Body, err)
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
		log.Printf("[FSM] Finish handle %s %v", msg.Body, err)
		if err != nil {
			if msg.Nack != nil {
				e := msg.Nack()
				log.Printf("[FSM] NACK %s %v", msg.Body, e)
			}
			return
		}
		if msg.Ack != nil {
			e := msg.Ack()
			log.Printf("[FSM] ACK %s %v", msg.Body, e)
		}
	}()
	log.Printf("[FSM] Start handle %s", msg.Body)

	c := msg.C
	taskID := msg.Body

	state, err := internal.QueryTaskState(c, w.GetDB(), w.Models, taskID)
	log.Printf("[FSM] Task %s %s %v", taskID, *state, err)
	if err != nil {
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

	return nil
}
