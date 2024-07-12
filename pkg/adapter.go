package pkg

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type IAdapter[Data DataEntity] interface {
	Init() error

	BeforeCreate(c context.Context, task *Task[Data]) error
	CreateCheck(c context.Context, task *Task[Data]) error
	Create(c context.Context, task *Task[Data]) error

	BeforeQuery(c context.Context, task *Task[Data]) error
	QueryCheck(c context.Context, task *Task[Data]) error
	Query(c context.Context, task *Task[Data]) error

	BeforeUpdate(c context.Context, task *Task[Data]) error
	UpdateCheck(c context.Context, task *Task[Data]) error
	Update(c context.Context, task *Task[Data]) error

	Publish(c context.Context, task *Task[Data]) error
}

type Adapter[Data DataEntity] struct {
	internal.Base[Data]
	ReInit         func() error
	ReBeforeCreate func(c context.Context, task *Task[Data]) error
	ReCreateCheck  func(c context.Context, task *Task[Data]) error
	ReCreate       func(c context.Context, task *Task[Data]) error
	ReBeforeQuery  func(c context.Context, task *Task[Data]) error
	ReQueryCheck   func(c context.Context, task *Task[Data]) error
	ReQuery        func(c context.Context, task *Task[Data]) error
	ReBeforeUpdate func(c context.Context, task *Task[Data]) error
	ReUpdateCheck  func(c context.Context, task *Task[Data]) error
	ReUpdate       func(c context.Context, task *Task[Data]) error
	RePublish      func(c context.Context, task *Task[Data]) error
}

func (a *Adapter[Data]) Init() error {
	if a.ReInit != nil {
		return a.ReInit()
	}

	if err := a.InitDB((*a.Config)[a.GetDBSection()].(util.Config)); err != nil {
		return err
	}

	if a.IMQ != nil {
		if err := a.InitMQ((*a.Config)[a.GetMQSection()].(util.Config)); err != nil {
			return err
		}
	}

	return nil
}

func (a *Adapter[Data]) BeforeCreate(c context.Context, task *Task[Data]) error {
	if a.ReBeforeCreate != nil {
		return a.ReBeforeCreate(c, task)
	}

	task.Version = 1
	return nil
}

func (a *Adapter[Data]) CreateCheck(c context.Context, task *Task[Data]) error {
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

func (a *Adapter[Data]) Create(c context.Context, task *Task[Data]) error {
	if a.ReCreate != nil {
		return a.ReCreate(c, task)
	}

	if err := a.BeforeCreate(c, task); err != nil {
		return err
	}
	if err := a.CreateCheck(c, task); err != nil {
		return err
	}

	task.SetTaskID(a.GenID())
	task.State = a.FSM.InitialState.GetName()

	if task.WithDB == nil {
		task.WithDB = a.GetDB()
	}
	if err := internal.CreateTask(c, a.Models, task); err != nil {
		return err
	}

	if err := a.Publish(c, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[Data]) BeforeQuery(c context.Context, task *Task[Data]) error {
	if a.ReBeforeQuery != nil {
		return a.ReBeforeQuery(c, task)
	}

	return nil
}

func (a *Adapter[Data]) QueryCheck(c context.Context, task *Task[Data]) error {
	if a.ReQueryCheck != nil {
		return a.ReQueryCheck(c, task)
	}

	if task.ID == "" && task.RequestID == "" {
		return fmt.Errorf("task.ID and task.RequestID both empty")
	}
	return nil
}

func (a *Adapter[Data]) Query(c context.Context, task *Task[Data]) error {
	if a.ReQuery != nil {
		return a.ReQuery(c, task)
	}

	if err := a.BeforeQuery(c, task); err != nil {
		return err
	}
	if err := a.QueryCheck(c, task); err != nil {
		return err
	}

	if task.WithDB == nil {
		task.WithDB = a.GetDB()
	}
	if err := internal.QueryTask(c, a.Models, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[Data]) BeforeUpdate(c context.Context, task *Task[Data]) error {
	if a.ReBeforeUpdate != nil {
		return a.ReBeforeUpdate(c, task)
	}

	return nil
}

func (a *Adapter[Data]) UpdateCheck(c context.Context, task *Task[Data]) error {
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

func (a *Adapter[Data]) Update(c context.Context, task *Task[Data]) error {
	if a.ReUpdate != nil {
		return a.ReUpdate(c, task)
	}

	if err := a.BeforeUpdate(c, task); err != nil {
		return err
	}
	if err := a.UpdateCheck(c, task); err != nil {
		return err
	}

	if task.WithDB == nil {
		task.WithDB = a.GetDB()
	}
	if err := internal.UpdateTask(c, a.Models, task, a.FSM); err != nil {
		return err
	}

	if err := a.Publish(c, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter[Data]) Publish(c context.Context, task *Task[Data]) error {
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
