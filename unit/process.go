package unit

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"github.com/stretchr/testify/mock"

	"ergo.services/testing/unit/stub"
)

type Process struct {
	*stub.Process

	priority    gen.MessagePriority
	important   bool
	node        *node
	callHelpers []CallHelper

	artifacts lib.QueueMPSC
}

func (p *Process) ValidateArtifacts(t testing.TB, expected []any) {
	t.Helper()

	if len(expected) == 0 {
		t.Fatal("ValidateArtifacts: no artifacts specified")
	}

	var actual []any
	for {
		a, ok := p.artifacts.Pop()
		if !ok {
			break
		}
		actual = append(actual, a)
	}

	if got, want := len(actual), len(expected); got != want {
		t.Fatalf("ValidateArtifacts: count mismatch: got %d, want %d", got, want)
	}

	if !reflect.DeepEqual(expected, actual) {
		// pretty-print to JSON
		expJSON, _ := marshalWithoutEscaping(expected)
		actJSON, _ := marshalWithoutEscaping(actual)

		expLines := strings.Split(expJSON, "\n")
		actLines := strings.Split(actJSON, "\n")

		var b strings.Builder
		max := len(expLines)
		if len(actLines) > max {
			max = len(actLines)
		}
		for i := 0; i < max; i++ {
			e, a := "", ""
			if i < len(expLines) {
				e = expLines[i]
			}
			if i < len(actLines) {
				a = actLines[i]
			}
			if e == a {
				b.WriteString("  " + e + "\n")
			} else {
				b.WriteString(fmt.Sprintf("- %s\n+ %s\n", e, a))
			}
		}

		t.Fatalf("ValidateArtifacts: mismatch (-expected +actual):\n%s", b.String())
	}
}

type processOptions struct {
	name        gen.Atom
	node        *node
	callHelpers []CallHelper
	priority    gen.MessagePriority
	important   bool
}

func newProcess(t testing.TB, artifacts lib.QueueMPSC, options processOptions) *Process {
	process := &Process{
		Process:     stub.NewProcess(t),
		artifacts:   artifacts,
		node:        options.node,
		callHelpers: options.callHelpers,
		priority:    options.priority,
		important:   options.important,
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

	process.On("SetSendPriority", mock.AnythingOfType("gen.MessagePriority")).
		Run(func(args mock.Arguments) {
			process.priority = args.Get(0).(gen.MessagePriority)
		}).
		Return(nil).
		Maybe()
	process.On("SendPriority").
		Return(func() gen.MessagePriority {
			return process.priority
		}).
		Maybe()

	process.On("SetImportantDelivery", mock.AnythingOfType("bool")).
		Run(func(args mock.Arguments) {
			process.important = args.Get(0).(bool)
		}).
		Return(nil).
		Maybe()
	process.On("ImportantDelivery").
		Return(func() bool {
			return process.important
		}).
		Maybe()

	// 1) the generic two-arg Send(dest, msg)
	closureSend := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  process.priority,
			Important: process.important,
		}
		process.artifacts.Push(art)
	}

	// 2) SendWithPriority(dest, msg, priority)
	closureSendWithPriority := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  args.Get(2).(gen.MessagePriority),
			Important: process.important,
		}
		process.artifacts.Push(art)
	}

	// 3) SendImportant(dest, msg)
	closureSendImportant := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  process.priority,
			Important: true,
		}
		process.artifacts.Push(art)
	}

	process.On("Send", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureSend).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureSend).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Atom"), mock.Anything).
		Run(closureSend).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureSend).Return(nil).Maybe()

	// SendWithPriority
	process.On("SendWithPriority", mock.Anything, mock.Anything, mock.AnythingOfType("gen.MessagePriority")).
		Run(closureSendWithPriority).
		Return(nil).
		Maybe()

	// SendImportant
	process.On("SendImportant", mock.Anything, mock.Anything).
		Run(closureSendImportant).
		Return(nil).
		Maybe()

	// SendAfter
	// Todo do we want to check gen.CancelFunc?
	process.On("SendAfter", mock.Anything, mock.Anything, mock.AnythingOfType("time.Duration")).
		Run(closureSend).
		Return(gen.CancelFunc(nil), nil).
		Maybe()

	process.On("SendEvent", mock.AnythingOfType("gen.Atom"), mock.AnythingOfType("gen.Ref"), mock.Anything).
		Run(func(args mock.Arguments) {
			art := ArtifactEvent{
				Name:     args.Get(0).(gen.Atom),
				Token:    args.Get(1).(gen.Ref),
				Message:  args.Get(2),
				Priority: process.priority,
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

	process.On("SendExit", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("error")).
		Run(func(args mock.Arguments) {
			art := ArtifactExit{
				To:     args.Get(0).(gen.PID),
				Reason: args.Get(1).(error),
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

	process.On("SendExitMeta", mock.AnythingOfType("gen.Alias"), mock.AnythingOfType("error")).
		Run(func(args mock.Arguments) {
			art := ArtifactExitMeta{
				Meta:   args.Get(0).(gen.Alias),
				Reason: args.Get(1).(error),
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

	process.On("SendResponse", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("gen.Ref"), mock.Anything).
		Run(func(args mock.Arguments) {
			art := ArtifactSend{
				From:     pid,
				To:       args.Get(0),
				Ref:      args.Get(1).(gen.Ref),
				Message:  args.Get(2),
				Priority: process.priority,
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

	process.On("SendResponseError", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("gen.Ref"), mock.AnythingOfType("error")).
		Run(func(args mock.Arguments) {
			art := ArtifactSend{
				From:     pid,
				To:       args.Get(0),
				Ref:      args.Get(1).(gen.Ref),
				Message:  args.Get(2).(error),
				Priority: process.priority,
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

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
