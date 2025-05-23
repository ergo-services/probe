package unit

import (
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
)

type SpawnOptions struct {
	Register          gen.Atom
	Parent            gen.PID
	Leader            gen.PID
	LogLevel          gen.LogLevel
	Priority          gen.MessagePriority
	ImportantDelivery bool

	Env map[gen.Env]any

	Helpers SpawnHelpers
}

type SpawnHelpers struct {
	Call []CallHelper
}

type CallHelper struct {
	Request  any
	Response any
}

func Spawn(t testing.TB, factory gen.ProcessFactory, options SpawnOptions, args ...any) (*Process, error) {
	behavior := factory()
	artifacts := lib.NewQueueMPSC()
	stubNode := newNode(t, artifacts)

	if options.LogLevel == gen.LogLevelDefault {
		options.LogLevel = gen.LogLevelTrace
	}
	stubNode.Log().SetLevel(options.LogLevel)

	popts := processOptions{
		node:         stubNode,
		SpawnOptions: options,
	}

	stubProcess := newProcess(t, artifacts, popts)
	stubProcess.On("Behavior").Return(behavior).Maybe()

	if err := behavior.ProcessInit(stubProcess, args...); err != nil {
		return nil, err
	}

	return stubProcess, nil
}
