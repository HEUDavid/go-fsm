package pkg

import (
	"context"
	"fmt"
	"github.com/HEUDavid/go-fsm/internal"
	. "github.com/HEUDavid/go-fsm/pkg/metadata"
	"github.com/HEUDavid/go-fsm/pkg/util"
)

type IAdapter interface {
	Init()

	BeforeCreate(c context.Context, task *Task[ExtDataEntity]) error
	CreateCheck(c context.Context, task *Task[ExtDataEntity]) error
	Create(c context.Context, task *Task[ExtDataEntity]) error

	BeforeQuery(c context.Context, task *Task[ExtDataEntity]) error
	QueryCheck(c context.Context, task *Task[ExtDataEntity]) error
	Query(c context.Context, task *Task[ExtDataEntity]) error

	BeforeUpdate(c context.Context, task *Task[ExtDataEntity]) error
	UpdateCheck(c context.Context, task *Task[ExtDataEntity]) error
	Update(c context.Context, task *Task[ExtDataEntity]) error

	Publish(c context.Context, task *Task[ExtDataEntity]) error
}

type Adapter struct {
	internal.Base
	IAdapter
}

func (a *Adapter) Init() {

	if err := a.InitDB((*a.Config)["mysql"].(util.Config)); err != nil {
		panic(err)
	}

	if a.IMQ != nil {
		if err := a.InitMQ((*a.Config)["rmq"].(util.Config)); err != nil {
			panic(err)
		}
	}

}

func (a *Adapter) BeforeCreate(c context.Context, task *Task[ExtDataEntity]) error {
	task.Version = 1
	return nil
}

func (a *Adapter) CreateCheck(c context.Context, task *Task[ExtDataEntity]) error {
	if task.RequestID == "" {
		return fmt.Errorf("task.RequestID empty")
	}
	if task.Type == "" {
		return fmt.Errorf("task.Type empty")
	}
	return nil
}

func (a *Adapter) Create(c context.Context, task *Task[ExtDataEntity]) error {

	if err := a.IAdapter.BeforeCreate(c, task); err != nil {
		return err
	}
	if err := a.IAdapter.CreateCheck(c, task); err != nil {
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

func (a *Adapter) BeforeQuery(c context.Context, task *Task[ExtDataEntity]) error {
	return nil
}

func (a *Adapter) QueryCheck(c context.Context, task *Task[ExtDataEntity]) error {
	if task.ID == "" && task.RequestID == "" {
		return fmt.Errorf("task.ID and task.RequestID both empty")
	}
	return nil
}

func (a *Adapter) Query(c context.Context, task *Task[ExtDataEntity]) error {

	if err := a.IAdapter.BeforeQuery(c, task); err != nil {
		return err
	}
	if err := a.IAdapter.QueryCheck(c, task); err != nil {
		return err
	}

	if err := internal.QueryTaskTx(c, a.GetDB(), a.TaskModel, a.ExtDataModel, task); err != nil {
		return err
	}

	return nil
}

func (a *Adapter) BeforeUpdate(c context.Context, task *Task[ExtDataEntity]) error {
	return nil
}

func (a *Adapter) UpdateCheck(c context.Context, task *Task[ExtDataEntity]) error {
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

func (a *Adapter) Update(c context.Context, task *Task[ExtDataEntity]) error {

	if err := a.IAdapter.BeforeUpdate(c, task); err != nil {
		return err
	}
	if err := a.IAdapter.UpdateCheck(c, task); err != nil {
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

func (a *Adapter) Publish(c context.Context, task *Task[ExtDataEntity]) error {
	if a.IMQ != nil {
		if err := a.PublishMessage(c, task.ID); err != nil {
			return err
		}
	}
	return nil
}
