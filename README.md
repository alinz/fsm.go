# fsm.go

fsm is a finite state machine implemented in Golang. This library was implemented as I needed a time base finite state machine to implement p2p network protocol.

## Installation

Since this library only requires a single file, I would suggest copy that file and place it under `internal/fsm` folder for your project. However if you don't want that you can use `go get` and add it under your `go.mod`,

```
go get github.com/alinz/fsm.go
```

## Usage

The best way to show the usage is by showing an example.

let's say we want to model a toggle button. There are 2 states a button can be at any time, `on` and `off`.

let's create a file call `button.go` under `button` folder

```go
package button

import (
	"github.com/alinz/fsm.go"
)

const (
	EvtToggle = fsm.Event("toggle") // define a toggle event
)

const (
	_ fsm.State = iota // we always ignore the first one as State can't be zero
	On  // store 1
	Off // store 2
)

func NewFSM() (*fsm.Machine, error) {
	m, err := fsm.NewMachine(fsm.Config{
		// define the initial state of machine, which in this case it would be Off
		Initial: off,
		States: fsm.States{
			// define the first state which is On
			{
				Ref: On,
				// define list of all events and their corresponding states
				On: fsm.On{
					{
						// once this machine receives EvtToggle, it
						// switch to Off state
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: Off,
							},
						},
					},
				},
			},
			// define the first state which is Off
			{
				Ref: Off,
				On: fsm.On{
					{
						// once this machine receives EvtToggle, it
						// switch to On state
						Event: EvtToggle,
						Targets: fsm.Targets{
							{
								Target: On,
							},
						},
					},
				},
			},
		},
	})

	return m, err
}
```

now in your main file,

```go
package main

import (
	"github.com/alinz/fsm.go"

	"example.com/internal/button"
)

func main() {
	machine, err := button.NewFSM()
	if err != nil {
		panic(err)
	}

	fmt.Println(machine.State()) // should print out initial state which is Off

	// let's send an event
	err = machine.Send(button.EvtToggle)
	// Send function might return `fms.ErrNoop` if the given event doesn't change the state
	// of the system
	if err != nil && err != fsm.ErrNoop {
		panic(err)
	}

	fmt.Println(machine.State()) // should print out On state
}

```

For more examples, please look into `fsm.test.go`
