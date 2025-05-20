package unit

import (
	"reflect"
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"github.com/stretchr/testify/mock"
)

type Process struct {
	*stub.Process

	node        *node
	callHelpers []CallHelper
	artifacts   lib.QueueMPSC
}

func (p *Process) ValidateArtifacts(t testing.TB, expected []any) {
	if len(expected) == 0 {
		t.Fatal("no artifacts specified")
	}

	if len(expected) != int(p.artifacts.Len()) {
		t.Fatalf("number of artifacts (%d) mismatch with the expecting ones (%d)", p.artifacts.Len(), len(expected))
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

type processOptions struct {
	name        gen.Atom
	node        *node
	callHelpers []CallHelper
}

func newProcess(t testing.TB, artifacts lib.QueueMPSC, options processOptions) *Process {
	process := &Process{
		Process:     stub.NewProcess(t),
		artifacts:   artifacts,
		node:        options.node,
		callHelpers: options.callHelpers,
	}
	nodeName := options.node.Name()
	creation := options.node.Creation()
	pid := gen.PID{Node: nodeName, ID: 1000, Creation: creation}

	process.On("Node").Return(options.node).Maybe()

	stubProcessLog := newStubLog(t, artifacts)
	stubProcessLog.SetLevel(options.node.Log().Level())

	process.On("Log").Return(stubProcessLog).Maybe()
	process.On("PID").Return(pid)
	process.On("Name").Return(options.name).Maybe()
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

	process.On("Call", mock.AnythingOfType("gen.PID"), mock.Anything).Return(func(to any, request any) (any, error) {
		art := ArtifactCall{
			From:    pid,
			To:      to,
			Request: request,
		}
		process.artifacts.Push(art)
		for _, helper := range process.callHelpers {
			if eq := reflect.DeepEqual(art.Request, helper.Request); eq == false {
				continue
			}
			return helper.Response, nil
		}

		return nil, nil
	}).Maybe()

	process.On("Call", mock.AnythingOfType("gen.ProcessID"), mock.Anything).Return(func(to any, request any) (any, error) {
		art := ArtifactCall{
			From:    pid,
			To:      to,
			Request: request,
		}
		process.artifacts.Push(art)
		for _, helper := range process.callHelpers {
			if eq := reflect.DeepEqual(art.Request, helper.Request); eq == false {
				continue
			}
			return helper.Response, nil
		}

		return nil, nil
	}).Maybe()
	process.On("Mailbox").Return(gen.ProcessMailbox{}).Maybe()

	closureCallTimeout := func(to any, request any, _ int) (any, error) {
		art := ArtifactCall{
			From:    pid,
			To:      to,
			Request: request,
		}
		process.artifacts.Push(art)
		for _, helper := range process.callHelpers {
			if eq := reflect.DeepEqual(art.Request, helper.Request); eq == false {
				continue
			}
			return helper.Response, nil
		}

		return nil, nil
	}
	closureCall := func(to any, request any) (any, error) {
		return closureCallTimeout(to, request, 1)
	}
	closureCallPID := func(to gen.PID, request any, timeout int) (any, error) {
		return closureCallTimeout(to, request, 1)
	}
	closureCallProcessID := func(to gen.ProcessID, request any, timeout int) (any, error) {
		return closureCallTimeout(to, request, 1)
	}
	closureCallAlias := func(to gen.Alias, request any, timeout int) (any, error) {
		return closureCallTimeout(to, request, 1)
	}

	process.On("Call", mock.AnythingOfType("gen.PID"), mock.Anything).Return(closureCall).Maybe()
	process.On("Call", mock.AnythingOfType("gen.Atom"), mock.Anything).Return(closureCall).Maybe()
	process.On("Call", mock.AnythingOfType("gen.ProcessID"), mock.Anything).Return(closureCall).Maybe()
	process.On("Call", mock.AnythingOfType("gen.Alias"), mock.Anything).Return(closureCall).Maybe()
	process.
		On("CallPID", mock.AnythingOfType("gen.PID"), mock.Anything, mock.AnythingOfType("int")).
		Return(closureCallPID).Maybe()
	process.
		On("CallProcessID", mock.AnythingOfType("gen.ProcessID"), mock.Anything, mock.AnythingOfType("int")).
		Return(closureCallProcessID).Maybe()
	process.
		On("CallAlias", mock.AnythingOfType("gen.Alias"), mock.Anything, mock.AnythingOfType("int")).
		Return(closureCallAlias).Maybe()

	process.On("Mailbox").Return(gen.ProcessMailbox{}).Maybe()
	return process
}
