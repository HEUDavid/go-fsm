package pkg

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type IAdapter[ExtData ExtDataEntity] interface {
	Init() error

	BeforeCreate(c context.Context, task *Task[ExtData]) error
	CreateCheck(c context.Context, task *Task[ExtData]) error
	Create(c context.Context, task *Task[ExtData]) error

	BeforeQuery(c context.Context, task *Task[ExtData]) error
	QueryCheck(c context.Context, task *Task[ExtData]) error
	Query(c context.Context, task *Task[ExtData]) error

	BeforeUpdate(c context.Context, task *Task[ExtData]) error
	UpdateCheck(c context.Context, task *Task[ExtData]) error
	Update(c context.Context, task *Task[ExtData]) error

	Publish(c context.Context, task *Task[ExtData]) error
}

type Adapter[ExtData ExtDataEntity] struct {
	internal.Base
	ReInit         func() error
	ReBeforeCreate func(c context.Context, task *Task[ExtData]) error
	ReCreateCheck  func(c context.Context, task *Task[ExtData]) error
	ReCreate       func(c context.Context, task *Task[ExtData]) error
	ReBeforeQuery  func(c context.Context, task *Task[ExtData]) error
	ReQueryCheck   func(c context.Context, task *Task[ExtData]) error
	ReQuery        func(c context.Context, task *Task[ExtData]) error
	ReBeforeUpdate func(c context.Context, task *Task[ExtData]) error
	ReUpdateCheck  func(c context.Context, task *Task[ExtData]) error
	ReUpdate       func(c context.Context, task *Task[ExtData]) error
	RePublish      func(c context.Context, task *Task[ExtData]) error
}

func (a *Adapter[ExtData]) Init() error {
	if a.ReInit != nil {
		return a.ReInit()
	}

	if err := a.InitDB((*a.Config)["mysql"].(util.Config)); err != nil {
		return err
	}

	if a.IMQ != nil {
		if err := a.InitMQ((*a.Config)["rmq"].(util.Config)); err != nil {
			return err
		}
	}

	return nil
}

func (a *Adapter[ExtData]) BeforeCreate(c context.Context, task *Task[ExtData]) error {
	if a.ReBeforeCreate != nil {
		return a.ReBeforeCreate(c, task)
	}

	task.Version = 1
	return nil
}

func (a *Adapter[ExtData]) CreateCheck(c context.Context, task *Task[ExtData]) error {
	if a.ReCreateCheck != nil {
		return a.ReCreateCheck(c, task)
	}

	if task.RequestID == "" {
		return fmt.Errorf("task.RequestID empty")
	}
	if task.Type == "" {
		return fmt.Errorf("task.Type empty")
	}
	return nil
}

func (a *Adapter[ExtData]) Create(c context.Context, task *Task[ExtData]) error {
	if a.ReCreate != nil {
		return a.ReCreate(c, task)
	}

	if err := a.BeforeCreate(c, task); err != nil {
		return err
	}
	if err := a.CreateCheck(c, task); err != nil {
		return err
	}

	task.SetTaskID(util.GenID())
	task.State = a.FSM.InitialState.GetName()

	if err := internal.CreateTaskTx(c, a.GetDB(), a.Models, task); err != nil {
		return err
	}

	if err := a.Publish(c, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[ExtData]) BeforeQuery(c context.Context, task *Task[ExtData]) error {
	if a.ReBeforeQuery != nil {
		return a.ReBeforeQuery(c, task)
	}

	return nil
}

func (a *Adapter[ExtData]) QueryCheck(c context.Context, task *Task[ExtData]) error {
	if a.ReQueryCheck != nil {
		return a.ReQueryCheck(c, task)
	}

	if task.ID == "" && task.RequestID == "" {
		return fmt.Errorf("task.ID and task.RequestID both empty")
	}
	return nil
}

func (a *Adapter[ExtData]) Query(c context.Context, task *Task[ExtData]) error {
	if a.ReQuery != nil {
		return a.ReQuery(c, task)
	}

	if err := a.BeforeQuery(c, task); err != nil {
		return err
	}
	if err := a.QueryCheck(c, task); err != nil {
		return err
	}

	if err := internal.QueryTaskTx(c, a.GetDB(), a.TaskModel, a.ExtDataModel, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[ExtData]) BeforeUpdate(c context.Context, task *Task[ExtData]) error {
	if a.ReBeforeUpdate != nil {
		return a.ReBeforeUpdate(c, task)
	}

	return nil
}

func (a *Adapter[ExtData]) UpdateCheck(c context.Context, task *Task[ExtData]) error {
	if a.ReUpdateCheck != nil {
		return a.ReUpdateCheck(c, task)
	}

	if task.RequestID == "" {
		return fmt.Errorf("task.RequestID empty")
	}
	if task.ID == "" {
		return fmt.Errorf("task.ID empty")
	}
	if task.Version <= 0 {
		return fmt.Errorf("task.Version empty")
	}
	return nil
}

func (a *Adapter[ExtData]) Update(c context.Context, task *Task[ExtData]) error {
	if a.ReUpdate != nil {
		return a.ReUpdate(c, task)
	}

	if err := a.BeforeUpdate(c, task); err != nil {
		return err
	}
	if err := a.UpdateCheck(c, task); err != nil {
		return err
	}

	if err := internal.UpdateTaskTx(c, a.GetDB(), a.Models, task, a.FSM); err != nil {
		return err
	}

	if err := a.Publish(c, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[ExtData]) Publish(c context.Context, task *Task[ExtData]) error {
	if a.RePublish != nil {
		return a.RePublish(c, task)
	}

	if a.IMQ != nil {
		if err := a.PublishMessage(c, task.ID); err != nil {
			return err
		}
	}
	return nil
}
