package pkg

import (
	"context"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
	"log"
)

type IWorker[Data DataEntity] interface {
	Init()
	Run()
	HandleMessage(c context.Context, msg string) error
}

type Worker[Data DataEntity] struct {
	internal.Base[Data]
}

func (w *Worker[Data]) Init() {
	if err := w.InitDB((*w.Config)["mysql"].(util.Config)); err != nil {
		panic(err)
	}

	if err := w.InitMQ((*w.Config)["rmq"].(util.Config)); err != nil {
		panic(err)
	}
	go w.Start()

}

func (w *Worker[Data]) Run() {
	go func() {
		for {
			c, msg, ack := w.FetchMessage(context.Background())
			err := w.HandleMessage(c, msg)
			if err == nil && ack != nil {
				_ = ack()
			}
		}
	}()
}

func (w *Worker[Data]) HandleMessage(c context.Context, msg string) error {
	log.Printf("[FSM] HandleMessage: %s", msg)

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

	if err = internal.QueryTaskTx(c, w.GetDB(), w.Models, task); err != nil {
		return err
	}

	if err = handler.Handle(task); err != nil {
		return err
	}

	task.RequestID = util.GenID()
	if err = internal.UpdateTaskTx(c, w.GetDB(), w.Models, task, w.FSM); err != nil {
		return err
	}

	if err = w.PublishMessage(c, taskID); err != nil {
		return err
	}

	return nil
}
