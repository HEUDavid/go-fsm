package pkg

import (
	"context"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type IWorker[ExtData ExtDataEntity] interface {
	Init()
	Run()
	HandleMsg(c context.Context, msg string) error
}

type Worker[ExtData ExtDataEntity] struct {
	internal.Base[ExtData]
}

func (w *Worker[ExtData]) Init() {
	if err := w.InitDB((*w.Config)["mysql"].(util.Config)); err != nil {
		panic(err)
	}

	if err := w.InitMQ((*w.Config)["rmq"].(util.Config)); err != nil {
		panic(err)
	}
	go w.Start()

}

func (w *Worker[ExtData]) Run() {
	go func() {
		for {
			c, msg, ack := w.FetchMessage(context.Background())
			err := w.HandleMsg(c, msg)
			if err == nil && ack != nil {
				_ = ack()
			}
		}
	}()
}

func (w *Worker[ExtData]) HandleMsg(c context.Context, msg string) error {

	taskID := msg
	state, err := internal.QueryTaskState(c, w.GetDB(), w.TaskModel, taskID)
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

	extData, _ := util.Assert[ExtData](util.ReflectNew(w.ExtDataModel))
	task := GenTaskInstance("", taskID, extData)

	if err = internal.QueryTaskTx(c, w.GetDB(), w.TaskModel, w.ExtDataModel, task); err != nil {
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
