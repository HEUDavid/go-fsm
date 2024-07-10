package metadata

import (
	"fmt"
)

type IState[Data DataEntity] interface {
	GetName() string
	IsFinalState() bool
	Handle(task *Task[Data]) error
}

type State[Data DataEntity] struct {
	Name    string
	IsFinal bool
	Handler func(task *Task[Data]) error
}

func (s State[Data]) GetName() string    { return s.Name }
func (s State[Data]) IsFinalState() bool { return s.IsFinal }
func (s State[Data]) Handle(task *Task[Data]) error {
	if s.Handler != nil {
		return s.Handler(task)
	}

	// TODO implement me
	panic("[FSM] implement me")
}

func GenState[Data DataEntity](name string, isFinal bool, handler func(task *Task[Data]) error) State[Data] {
	return State[Data]{name, isFinal, handler}
}

type ITransition[Data DataEntity] interface {
	GetName() string
}

type Transition[Data DataEntity] struct {
	From State[Data]
	To   State[Data]
}

func (t Transition[Data]) GetName() string {
	return fmt.Sprintf("%s->%s", t.From.GetName(), t.To.GetName())
}

func GenTransition[Data DataEntity](from, to State[Data]) Transition[Data] {
	return Transition[Data]{from, to}
}

type IFSM[Data DataEntity] interface {
	RegisterState(states ...State[Data])
	RegisterTransition(transitions ...Transition[Data])
	GetState(state string) (State[Data], bool)
	GetTransition(fromState, toState string) (Transition[Data], bool)
}

type FSM[Data DataEntity] struct {
	InitialState State[Data]
	States       map[string]State[Data]
	Transitions  map[string]Transition[Data]
}

func (f *FSM[Data]) GetState(state string) (State[Data], bool) {
	s, exist := f.States[state]
	return s, exist
}

func (f *FSM[Data]) GetTransition(fromState, toState string) (Transition[Data], bool) {
	t, exist := f.Transitions[fmt.Sprintf("%s->%s", fromState, toState)]
	return t, exist
}

func (f *FSM[Data]) RegisterState(states ...State[Data]) {
	for _, state := range states {
		f.States[state.GetName()] = state
	}
}

func (f *FSM[Data]) RegisterTransition(transitions ...Transition[Data]) {
	for _, transition := range transitions {
		f.Transitions[transition.GetName()] = transition
	}
}

func GenFSM[Data DataEntity](initial State[Data]) FSM[Data] {
	return FSM[Data]{
		InitialState: initial,
		States:       map[string]State[Data]{},
		Transitions:  map[string]Transition[Data]{},
	}
}
