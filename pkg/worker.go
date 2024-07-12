package pkg

import (
	"context"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/mq"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"log"
	"sync"
)

type IWorker[Data DataEntity] interface {
	Init()
	Run()
	Handle(c context.Context, msg string, ack mq.ACK) error
}

type Worker[Data DataEntity] struct {
	internal.Base[Data]
	MaxGoroutines int
	ReInit        func()
	ReRun         func()
	ReHandle      func(c context.Context, msg string, ack mq.ACK) error
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
	go w.Start()
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

				c, msg, ack := w.FetchMessage(context.Background())
				if err := w.Handle(c, msg, ack); err != nil {
					log.Printf("[FSM] Handle Err: %v", err)
				}
			}()
		}
	}()
}

func (w *Worker[Data]) Handle(c context.Context, msg string, ack mq.ACK) (err error) {
	if w.ReHandle != nil {
		return w.ReHandle(c, msg, ack)
	}

	defer func() {
		log.Printf("[FSM] Finish handle %s, %v", msg, err)
		if err != nil {
			return
		}
		err = ack()
		log.Printf("[FSM] ACK %s, %v", msg, err)
	}()
	log.Printf("[FSM] Start handle %s", msg)

	taskID := msg
	state, err := internal.QueryTaskState(c, w.GetDB(), w.Models, taskID)
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
