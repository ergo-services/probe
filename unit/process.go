package unit

import (
	"reflect"
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/probe/unit/stub"
	"github.com/stretchr/testify/mock"
)

type Process struct {
	*stub.Process

	artifacts lib.QueueMPSC
}

func (p *Process) ValidateArtifacts(t testing.TB, expected []any) {
	if len(expected) != int(p.artifacts.Len()) {
		t.Fatal("number of artifacts mismatch with the expecting ones")
	}

	for {
		a, exist := p.artifacts.Pop()
		if exist == false {
			return
		}

		if equal := reflect.DeepEqual(a, expected[0]); equal == false {
			t.Fatalf("expected artifact %#v mismatch: %#v", expected[0], a)
		}
		expected = expected[1:]
	}
}

func newProcess(t testing.TB, artifacts lib.QueueMPSC, name gen.Atom, node *node) *Process {
	process := &Process{
		Process:   stub.NewProcess(t),
		artifacts: artifacts,
	}
	nodeName := node.Name()
	creation := node.Creation()
	pid := gen.PID{Node: nodeName, ID: 1000, Creation: creation}

	process.On("Node").Return(node).Maybe()

	stubProcessLog := newStubLog(t, artifacts)
	stubProcessLog.SetLevel(node.Log().Level())

	process.On("Log").Return(stubProcessLog).Maybe()
	process.On("PID").Return(pid)
	process.On("Name").Return(name).Maybe()
	process.On("Send", mock.AnythingOfType("gen.PID"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Message: args.Get(1),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.ProcessID"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Message: args.Get(1),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Atom"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Message: args.Get(1),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Alias"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Message: args.Get(1),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("Mailbox").Return(gen.ProcessMailbox{}).Maybe()

	return process
}
