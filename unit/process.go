package unit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"github.com/stretchr/testify/mock"

	"ergo.services/probe/unit/stub"
)

type Process struct {
	*stub.Process

	Priority  gen.MessagePriority
	Important bool

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
		expJSON, _ := json.MarshalIndent(expected, "", "  ")
		actJSON, _ := json.MarshalIndent(actual, "", "  ")
		expLines := strings.Split(string(expJSON), "\n")
		actLines := strings.Split(string(actJSON), "\n")

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

	process.On("SetSendPriority", mock.AnythingOfType("gen.MessagePriority")).Run(func(args mock.Arguments) {
		process.Priority = args.Get(0).(gen.MessagePriority)
	}).Return(nil).Maybe()
	process.On("SendPriority").Return(func() gen.MessagePriority {
		return process.Priority
	}).Maybe()

	process.On("SetImportantDelivery", mock.AnythingOfType("bool")).Run(func(args mock.Arguments) {
		process.Important = args.Get(0).(bool)
	}).Return(nil).Maybe()

	process.On("ImportantDelivery").Return(func() bool {
		return process.Important
	}).Maybe()

	process.On("Send", mock.AnythingOfType("gen.PID"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     pid,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: process.Priority,
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.ProcessID"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     pid,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: process.Priority,
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Atom"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     pid,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: process.Priority,
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()
	process.On("Send", mock.AnythingOfType("gen.Alias"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     pid,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: process.Priority,
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendWithPriority", mock.Anything, mock.Anything, mock.AnythingOfType("gen.MessagePriority")).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     pid,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: args.Get(2).(gen.MessagePriority),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendImportant", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Important: true,
			Priority:  process.Priority,
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	// Todo do we want to check gen.CancelFunc?
	process.On("SendAfter", mock.Anything, mock.Anything, mock.AnythingOfType("time.Duration")).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Message: args.Get(1),
		}
		process.artifacts.Push(art)
	}).Return(gen.CancelFunc(nil), nil).Maybe()

	process.On("SendEvent", mock.AnythingOfType("gen.Atom"), mock.AnythingOfType("gen.Ref"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactEvent{
			Name:    args.Get(0).(gen.Atom),
			Token:   args.Get(1).(gen.Ref),
			Message: args.Get(2),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendExit", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("error")).Run(func(args mock.Arguments) {
		art := ArtifactExit{
			To:     args.Get(0).(gen.PID),
			Reason: args.Get(1).(error),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendExitMeta", mock.AnythingOfType("gen.Alias"), mock.AnythingOfType("error")).Run(func(args mock.Arguments) {
		art := ArtifactExitMeta{
			Meta:   args.Get(0).(gen.Alias),
			Reason: args.Get(1).(error),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendResponse", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("gen.Ref"), mock.Anything).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Ref:     args.Get(1).(gen.Ref),
			Message: args.Get(2),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("SendResponseError", mock.AnythingOfType("gen.PID"), mock.AnythingOfType("gen.Ref"), mock.AnythingOfType("error")).Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:    pid,
			To:      args.Get(0),
			Ref:     args.Get(1).(gen.Ref),
			Message: args.Get(2).(error),
		}
		process.artifacts.Push(art)
	}).Return(nil).Maybe()

	process.On("Mailbox").Return(gen.ProcessMailbox{}).Maybe()

	return process
}
