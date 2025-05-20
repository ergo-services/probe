package unit

import (
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
)

type SpawnOptions struct {
	LogLevel          gen.LogLevel
	Priority          gen.MessagePriority
	ImportantDelivery bool
}

func Spawn(t testing.TB, factory gen.ProcessFactory, options SpawnOptions, args ...any) (*Process, error) {
	return SpawnRegister(t, "", factory, options, args...)
}

func SpawnRegister(t testing.TB, name gen.Atom, factory gen.ProcessFactory, options SpawnOptions, args ...any) (*Process, error) {

	behavior := factory()
	artifacts := lib.NewQueueMPSC()
	stubNode := newNode(t, artifacts)

	stubNode.Log().SetLevel(options.LogLevel)

	stubProcess := newProcess(t, artifacts, name, stubNode)
	_ = stubProcess.SetSendPriority(options.Priority)
	_ = stubProcess.SetImportantDelivery(options.ImportantDelivery)
	stubProcess.On("Behavior").Return(behavior).Maybe()
	err := behavior.ProcessInit(stubProcess, args...)
	return stubProcess, err
}
