package fsm_test

import (
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

	m, err := fsm.NewMachine(fsm.Config{
		Initial: off,
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

	m, err := fsm.NewMachine(fsm.Config{
		Initial: red,
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

	testCases := []struct {
		description   string
		wait          time.Duration
		expectedState fsm.State
	}{
		{
			description:   "change from red to green",
			wait:          750 * time.Millisecond,
			expectedState: green,
		},
		{
			description:   "change from green to yellow",
			wait:          750 * time.Millisecond,
			expectedState: yellow,
		},
		{
			description:   "change from yellow to red",
			wait:          750 * time.Millisecond,
			expectedState: green,
		},
	}

	for _, testCase := range testCases {
		time.Sleep(testCase.wait)

		if m.State() != testCase.expectedState {
			t.Errorf("in %s, expected %s but got %s", testCase.description, state2String(testCase.expectedState), state2String(m.State()))
		}
	}

}
