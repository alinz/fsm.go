package fsm

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInitialNotSet happens when Initial value not being set in Machine's config
	ErrInitialNotSet = errors.New("initial state is required")
	// ErrDuplicateState happens when an state defines more than once
	ErrDuplicateState = errors.New("state is duplicated")
	// ErrNoop happens when state doesn't change upon calling Send method
	ErrNoop = errors.New("no change")
	// ErrCondFailed happens at Send and initial moment if Cond fails
	ErrCondFailed = errors.New("condition failed")
	// ErrStateNotFound happens when an unknown state is being set
	ErrStateNotFound = errors.New("state not found")
)

// Event is a custom type which defines machine's events
type Event string

// State is a custom type which defines machine's states
type State uint32

// Timeout is part of configuration which defines a timeout
// once the Duration is passed, machines tries to change to
// one of the given states at On field
type Timeout struct {
	Duration time.Duration
	Targets  Targets
}

// States list of all state's
type States []struct {
	Ref     State
	Timeout *Timeout
	On      On
}

// Targets defines the next state, if Cond is defined, first it checks the Cond upon moving to state
type Targets []struct {
	Cond   func() bool
	Target State
}

// On defines all states related to given State
type On []struct {
	Event   Event
	Cond    func() bool
	Targets Targets
}

// Config defines the Machine's configuration
type Config struct {
	Initial State
	States  States
}

type key struct {
	Ref   State
	Event Event
}

type stateInfo struct {
	Timeout *Timeout
}

type stateEventInfo struct {
	Cond    func() bool
	Targets Targets
}

type Machine struct {
	currentState  State
	states        map[State]*stateInfo
	nextStates    map[key]*stateEventInfo
	cancelTimeout func()
}

// Send sends an event to machine, if nothing changes, ErrNoop will be return
func (m *Machine) Send(evt Event) error {
	key := key{m.currentState, evt}
	stateEventInfo, ok := m.nextStates[key]
	if !ok {
		return ErrNoop
	}

	if stateEventInfo.Cond != nil && !stateEventInfo.Cond() {
		return ErrCondFailed
	}

	for _, target := range stateEventInfo.Targets {
		if target.Cond != nil && !target.Cond() {
			continue
		}

		return m.process(target.Target)
	}

	return ErrNoop
}

func (m *Machine) process(state State) error {
	if m.cancelTimeout != nil {
		m.cancelTimeout()
		m.cancelTimeout = nil
	}

	stateInfo, ok := m.states[state]
	if !ok {
		return ErrStateNotFound
	}

	if stateInfo.Timeout == nil {
		// No timeout set, simply assing target to current
		m.currentState = state
		return nil
	}

	// need to setup timeout
	m.cancelTimeout = setTimeout(func() {
		for _, state := range stateInfo.Timeout.Targets {
			if state.Cond != nil && !state.Cond() {
				continue
			}

			m.currentState = state.Target
			m.process(m.currentState)
			break
		}
	}, stateInfo.Timeout.Duration)

	return nil
}

// State returns the current state of machine
func (m Machine) State() State {
	return m.currentState
}

// NewMachine creates a new machine
func NewMachine(conf Config) (*Machine, error) {
	if conf.Initial == 0 {
		return nil, ErrInitialNotSet
	}

	states := make(map[State]*stateInfo)
	nextStates := make(map[key]*stateEventInfo)

	for _, state := range conf.States {
		if _, ok := states[state.Ref]; ok {
			return nil, fmt.Errorf("duplicate state ref %d: %w", state.Ref, ErrDuplicateState)
		}

		for _, nextState := range state.On {
			nextStates[key{state.Ref, nextState.Event}] = &stateEventInfo{
				Cond:    nextState.Cond,
				Targets: nextState.Targets,
			}
		}

		states[state.Ref] = &stateInfo{
			Timeout: state.Timeout,
		}
	}

	m := &Machine{
		currentState: conf.Initial,
		nextStates:   nextStates,
		states:       states,
	}

	err := m.process(conf.Initial)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func setTimeout(fn func(), timeout time.Duration) func() {
	cancel := make(chan struct{}, 1)

	go func() {
		select {
		case <-time.After(timeout):
			fn()
		case <-cancel:
		}
	}()

	return func() {
		close(cancel)
	}
}
