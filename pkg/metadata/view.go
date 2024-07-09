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
	Name    string
	IsFinal bool
	Handler func(task *Task[ExtData]) error
}

func (s State[ExtData]) GetName() string    { return s.Name }
func (s State[ExtData]) IsFinalState() bool { return s.IsFinal }
func (s State[ExtData]) Handle(task *Task[ExtData]) error {
	if s.Handler != nil {
		return s.Handler(task)
	}

	// TODO implement me
	panic("implement me")
}

func GenState[ExtData ExtDataEntity](name string, isFinal bool, handler func(task *Task[ExtData]) error) State[ExtData] {
	return State[ExtData]{
		Name:    name,
		IsFinal: isFinal,
		Handler: handler,
	}
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

func GenTransition[ExtData ExtDataEntity](from, to State[ExtData]) Transition[ExtData] {
	return Transition[ExtData]{from, to}
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

func (f *FSM[ExtData]) GetState(stateName string) (State[ExtData], bool) {
	state, exist := f.States[stateName]
	return state, exist
}

func (f *FSM[ExtData]) GetTransition(from, to string) (Transition[ExtData], bool) {
	transition, exist := f.Transitions[fmt.Sprintf("%s->%s", from, to)]
	return transition, exist
}

func (f *FSM[ExtData]) RegisterState(states ...State[ExtData]) {
	for _, state := range states {
		f.States[state.GetName()] = state
	}
}

func (f *FSM[ExtData]) RegisterTransition(transitions ...Transition[ExtData]) {
	for _, transition := range transitions {
		f.Transitions[transition.GetName()] = transition
	}
}

func GenFSM[ExtData ExtDataEntity](Initial State[ExtData]) FSM[ExtData] {
	return FSM[ExtData]{
		InitialState: Initial,
		States:       map[string]State[ExtData]{},
		Transitions:  map[string]Transition[ExtData]{},
	}
}
