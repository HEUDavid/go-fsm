package metadata

import (
	"fmt"
)

type IState interface {
	GetName() string
	IsFinalState() bool
	Handle(task *Task[ExtDataEntity]) error
}

type State struct {
	IState
	Name    string
	IsFinal bool
}

func (s State) GetName() string                        { return s.Name }
func (s State) IsFinalState() bool                     { return s.IsFinal }
func (s State) Handle(task *Task[ExtDataEntity]) error { return nil }

type ITransition interface {
	GetName() string
}

type Transition struct {
	From IState
	To   IState
}

func (t Transition) GetName() string {
	return fmt.Sprintf("%s->%s", t.From.GetName(), t.To.GetName())
}

type IFSM interface {
	GetState(stateName string) (State, bool)
	GetTransition(fromState, toState string) (*Transition, bool)
}

type FSM struct {
	InitialState IState
	States       map[string]IState
	Transitions  map[string]Transition
}

func (f FSM) GetState(stateName string) (IState, bool) {
	state, exist := f.States[stateName]
	return state, exist
}

func (f FSM) GetTransition(fromState, toState string) (Transition, bool) {
	transition, exist := f.Transitions[fmt.Sprintf("%s->%s", fromState, toState)]
	return transition, exist
}
