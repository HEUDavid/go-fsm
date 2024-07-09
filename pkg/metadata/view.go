package metadata

import (
	"fmt"
)

type IState[ExtData ExtDataEntity] interface {
	GetName() string
	IsFinalState() bool
	Handle(task *Task[ExtData]) error
}

type State[ExtData ExtDataEntity] struct {
	Name     string
	IsFinal  bool
	ReHandle func(task *Task[ExtData]) error
}

func (s State[ExtData]) GetName() string    { return s.Name }
func (s State[ExtData]) IsFinalState() bool { return s.IsFinal }
func (s State[ExtData]) Handle(task *Task[ExtData]) error {
	if s.ReHandle != nil {
		return s.ReHandle(task)
	}

	// TODO implement me
	panic("implement me")
}

type ITransition[ExtData ExtDataEntity] interface {
	GetName() string
}

type Transition[ExtData ExtDataEntity] struct {
	From State[ExtData]
	To   State[ExtData]
}

func (t Transition[ExtData]) GetName() string {
	return fmt.Sprintf("%s->%s", t.From.GetName(), t.To.GetName())
}

type IFSM[ExtData ExtDataEntity] interface {
	GetState(stateName string) (State[ExtData], bool)
	GetTransition(fromState, toState string) (Transition[ExtData], bool)
}

type FSM[ExtData ExtDataEntity] struct {
	InitialState State[ExtData]
	States       map[string]State[ExtData]
	Transitions  map[string]Transition[ExtData]
}

func (f FSM[ExtData]) GetState(stateName string) (State[ExtData], bool) {
	state, exist := f.States[stateName]
	return state, exist
}

func (f FSM[ExtData]) GetTransition(from, to string) (Transition[ExtData], bool) {
	transition, exist := f.Transitions[fmt.Sprintf("%s->%s", from, to)]
	return transition, exist
}
