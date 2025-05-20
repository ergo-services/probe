package unit

import (
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
)

type CallHelper struct {
	Request  any
	Response any
}
type SpawnOptions struct {
	LogLevel          gen.LogLevel
	Priority          gen.MessagePriority
	ImportantDelivery bool
	CallHelpers       []CallHelper
}

func Spawn(t testing.TB, factory gen.ProcessFactory, options SpawnOptions, args ...any) (*Process, error) {
	return SpawnRegister(t, "", factory, options, args...)
}

func SpawnRegister(t testing.TB, name gen.Atom, factory gen.ProcessFactory, options SpawnOptions, args ...any) (*Process, error) {

	behavior := factory()
	artifacts := lib.NewQueueMPSC()
	stubNode := newNode(t, artifacts)

	stubNode.Log().SetLevel(options.LogLevel)

	popts := processOptions{
		name:        name,
		node:        stubNode,
		callHelpers: options.CallHelpers,
		priority:    options.Priority,
		important:   options.ImportantDelivery,
	}
	stubProcess := newProcess(t, artifacts, popts)
	stubProcess.On("Behavior").Return(behavior).Maybe()
	if err := behavior.ProcessInit(stubProcess, args...); err != nil {
		return nil, err
	}
	return stubProcess, nil
}
