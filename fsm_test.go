package fsm_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alinz/fsm.go"
)

func TestSimpleToggleMachine(t *testing.T) {

	const (
		EvtToggle = fsm.Event("toggle")
	)

	const (
		_ fsm.State = iota
		on
		off
	)

	state2String := func(state fsm.State) string {
		switch state {
		case on:
			return "on"
		case off:
			return "off"
		default:
			return "unknown state"
		}
	}

	m, err := fsm.NewMachine(fsm.Config{
		Initial: off,
		StateChanged: func(prev fsm.State, next fsm.State) {
			fmt.Printf("%s -> %s\n", state2String(prev), state2String(next))
		},
		States: fsm.States{
			{
				Ref: on,
				On: fsm.On{
					{
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: off,
							},
						},
					},
				},
			},
			{
				Ref: off,
				On: fsm.On{
					{
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: on,
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Errorf("failed to initialized machine: %s", err)
	}

	if m.State() != off {
		t.Errorf("initial state is not correctly set")
	}

	testCases := []struct {
		description   string
		event         fsm.Event
		sendError     error
		expectedState fsm.State
	}{
		{
			description:   "changing state from off to on",
			event:         EvtToggle,
			sendError:     nil,
			expectedState: on,
		},
		{
			description:   "changing state from on to off",
			event:         EvtToggle,
			sendError:     nil,
			expectedState: off,
		},
		{
			description:   "changing state from on to off",
			event:         "",
			sendError:     fsm.ErrNoop,
			expectedState: off,
		},
	}

	for _, testCase := range testCases {
		err = m.Send(testCase.event)
		if err != testCase.sendError {
			t.Errorf("in %s, expect to %s, but got %s error", testCase.description, testCase.sendError, err)
		}

		if m.State() != testCase.expectedState {
			t.Errorf("in %s, expected %d state but got %d", testCase.description, testCase.expectedState, m.State())
		}
	}
}

func TestTrafficLightMachine(t *testing.T) {
	const (
		EvtToggle = fsm.Event("toggle")
	)

	const (
		_      fsm.State = iota
		red              // 1
		yellow           // 2
		green            // 3
	)

	state2String := func(state fsm.State) string {
		switch state {
		case red:
			return "red"
		case yellow:
			return "yellow"
		case green:
			return "green"
		default:
			return "unknown"
		}
	}

	var wg sync.WaitGroup

	wg.Add(5)

	result := make([]string, 0)

	m, err := fsm.NewMachine(fsm.Config{
		Initial: red,
		StateChanged: func(prev fsm.State, next fsm.State) {
			result = append(result, fmt.Sprintf("%s->%s", state2String(prev), state2String((next))))
			wg.Done()
		},
		States: fsm.States{
			{
				Ref: red,
				Timeout: &fsm.Timeout{
					Duration: 500 * time.Millisecond,
					Targets: fsm.Targets{
						{
							Target: green,
						},
					},
				},
				On: fsm.On{
					{
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: green,
							},
						},
					},
				},
			},
			{
				Ref: yellow,
				Timeout: &fsm.Timeout{
					Duration: 500 * time.Millisecond,
					Targets: fsm.Targets{
						{
							Target: red,
						},
					},
				},
				On: fsm.On{
					{
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: red,
							},
						},
					},
				},
			},
			{
				Ref: green,
				Timeout: &fsm.Timeout{
					Duration: 500 * time.Millisecond,
					Targets: fsm.Targets{
						{
							Target: yellow,
						},
					},
				},
				On: fsm.On{
					{
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: yellow,
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Errorf("failed to initialized machine: %s", err)
	}

	if m.State() != red {
		t.Errorf("initial state is not correctly set")
	}

	wg.Wait()

	expected := []string{
		"red->green",
		"green->yellow",
		"yellow->red",
		"red->green",
		"green->yellow",
	}

	for i, value := range expected {
		if result[i] != value {
			t.Errorf("expected %s, but got %s at %d iteration", value, result[i], i)
		}
	}

}

// For the actual represtation of this state machine
// please see this URL
// https://excalidraw.com/#json=6233155535110144,NJZ-TsUF-K-rL8OLkCiCFA
func TestExampleDoor(t *testing.T) {
	const (
		_ fsm.State = iota
		Unlocked
		Closed
		Opened
		Locked
	)

	const (
		EvtOpen   = fsm.Event("open")
		EvtClose  = fsm.Event("close")
		EvtLock   = fsm.Event("lock")
		EvtUnlock = fsm.Event("unlock")
	)

	door, err := fsm.NewMachine(fsm.Config{
		Initial: Closed,
		States: fsm.States{
			{
				Ref: Closed,
				Timeout: &fsm.Timeout{
					Duration: 10 * time.Second,
					Targets: fsm.Targets{
						{
							Target: Locked,
						},
					},
				},
				On: fsm.On{
					{
						Event: EvtLock,
						Targets: fsm.Targets{
							{
								Target: Locked,
							},
						},
					},
					{
						Event: EvtOpen,
						Targets: fsm.Targets{
							{
								Target: Opened,
							},
						},
					},
				},
			},
			{
				Ref: Locked,
				On: fsm.On{
					{
						Event: EvtUnlock,
						Targets: fsm.Targets{
							{
								Target: Unlocked,
							},
						},
					},
				},
			},
			{
				Ref: Unlocked,
				On: fsm.On{
					{
						Event: EvtOpen,
						Targets: fsm.Targets{
							{
								Target: Opened,
							},
						},
					},
					{
						Event: EvtLock,
						Targets: fsm.Targets{
							{
								Target: Locked,
							},
						},
					},
				},
			},
			{
				Ref: Opened,
				On: fsm.On{
					{
						Event: EvtClose,
						Targets: fsm.Targets{
							{
								Target: Closed,
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		t.Errorf("failed to create door fsm: %s", err)
	}

	_ = door
}
