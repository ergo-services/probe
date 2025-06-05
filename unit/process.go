package unit

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/mock"
)

type Process struct {
	*stub.Process

	options   processOptions
	artifacts lib.QueueMPSC
	uniq      uint64
}

func (p *Process) ValidateArtifacts(t testing.TB, expected []any) (left int) {
	t.Helper()

	// slice must have the same capacity
	actual := make([]any, 0, len(expected))
	for {
		a, ok := p.artifacts.Pop()
		if !ok {
			if len(expected) == 0 {
				return 0
			}

			t.Fatalf("ValidateArtifacts: count mismatch: got %d, want %d",
				len(actual), len(expected))
		}

		left = int(p.artifacts.Len())

		actual = append(actual, a)
		if len(actual) < len(expected) {
			continue
		}

		break
	}

	if !reflect.DeepEqual(expected, actual) {
		// pretty-print
		expectedPP := spew.Sdump(expected)
		actualPP := spew.Sdump(actual)

		expLines := strings.Split(expectedPP, "\n")
		actLines := strings.Split(actualPP, "\n")

		var b strings.Builder
		max := len(expLines)
		if len(actLines) > max {
			max = len(actLines)
		}

		stringEqual := true
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
				stringEqual = false
				b.WriteString(fmt.Sprintf("- %s\n+ %s\n", e, a))
			}
		}

		// DeepEqual doesnt work for closures (gen.ProcessFactory as an example)
		if stringEqual == false {
			t.Fatalf("ValidateArtifacts: mismatch (-expected +actual):\n%s", b.String())
		}
	}

	return
}

func (p *Process) ResetArtifacts() int {
	l := p.artifacts.Len()
	for {
		_, ok := p.artifacts.Pop()
		if !ok {
			break
		}
	}
	return int(l)
}

type processOptions struct {
	SpawnOptions
	node *node
}

func newProcess(t testing.TB, artifacts lib.QueueMPSC, options processOptions) *Process {
	process := &Process{
		Process:   stub.NewProcess(t),
		artifacts: artifacts,
		options:   options,
	}

	nodeName := options.node.Name()
	creation := options.node.Creation()
	process.uniq = 1000
	pid := gen.PID{Node: nodeName, ID: 1000, Creation: creation}

	emptyPID := gen.PID{}
	if options.Parent == emptyPID {
		options.Parent = options.node.PID()
	}
	process.On("Parent").Return(options.Parent).Maybe()

	if options.Leader == emptyPID {
		options.Leader = options.node.PID()
	}
	process.On("Leader").Return(options.Leader).Maybe()

	process.On("Node").Return(options.node).Maybe()

	stubProcessLog := newStubLog(t, artifacts)
	stubProcessLog.SetLevel(options.node.Log().Level())

	process.On("Log").Return(stubProcessLog).Maybe()
	process.On("PID").Return(pid).Maybe()

	process.On("Env", mock.AnythingOfType("gen.Env")).Return(func(name gen.Env) (any, bool) {
		v, found := process.options.Env[name]
		return v, found
	}).Maybe()

	process.On("EnvDefault", mock.AnythingOfType("gen.Env"), mock.Anything).Return(func(name gen.Env, def any) any {
		v, found := process.options.Env[name]
		if found == false {
			return def
		}
		return v
	}).Maybe()

	process.
		On("SetEnv", mock.AnythingOfType("gen.Env"), mock.Anything).
		Run(func(args mock.Arguments) {
			if process.options.Env == nil {
				process.options.Env = make(map[gen.Env]any)
			}
			process.options.Env[args.Get(0).(gen.Env)] = args.Get(1)
		}).Maybe()

	process.
		On(
			"Spawn",
			mock.AnythingOfType("gen.ProcessFactory"),
			mock.AnythingOfType("gen.ProcessOptions"),
			mock.Anything,
		).
		Run(func(args mock.Arguments) {
			art := ArtifactSpawn{
				Factory: args.Get(0).(gen.ProcessFactory),
				Options: args.Get(1).(gen.ProcessOptions),
			}
			if len(args) == 3 {
				art.Args = args.Get(2).([]any)
			}
			process.uniq++
			process.artifacts.Push(art)
		}).Return(gen.PID{Node: nodeName, ID: process.uniq, Creation: creation}, nil).Maybe()

	process.
		On(
			"SpawnMeta",
			mock.MatchedBy(func(arg any) bool {
				_, ok := arg.(gen.MetaBehavior)
				return ok
			}),
			mock.AnythingOfType("gen.MetaOptions"),
		).
		Run(func(args mock.Arguments) {
			art := ArtifactSpawnMeta{
				Factory: args.Get(0).(gen.MetaBehavior),
				Options: args.Get(1).(gen.MetaOptions),
			}
			process.uniq++
			process.artifacts.Push(art)
		}).Return(gen.Alias{Node: nodeName, ID: [3]uint64{process.uniq, 0, 0}, Creation: creation}, nil).Maybe()

	process.On("Name").Return(func() gen.Atom {
		return process.options.Register
	}).Maybe()

	process.On("RegisterName", mock.AnythingOfType("gen.Atom")).
		Return(func(name gen.Atom) error {
			process.options.Register = name
			return nil
		}).Maybe()
	process.On("UnregisterName", mock.AnythingOfType("gen.Atom")).
		Return(func(name gen.Atom) error {
			process.options.Register = ""
			return nil
		}).Maybe()

	process.On("Uptime").Return(0).Maybe()

	process.On("SetSendPriority", mock.AnythingOfType("gen.MessagePriority")).
		Run(func(args mock.Arguments) {
			process.options.Priority = args.Get(0).(gen.MessagePriority)
		}).
		Return(nil).
		Maybe()
	process.On("SendPriority").
		Return(func() gen.MessagePriority {
			return process.options.Priority
		}).
		Maybe()

	process.On("SetImportantDelivery", mock.AnythingOfType("bool")).
		Run(func(args mock.Arguments) {
			process.options.ImportantDelivery = args.Get(0).(bool)
		}).
		Return(nil).
		Maybe()
	process.On("ImportantDelivery").
		Return(func() bool {
			return process.options.ImportantDelivery
		}).
		Maybe()

	process.On("State").
		Return(gen.ProcessStateRunning).Maybe()

	process.On("Mailbox").
		Return(gen.ProcessMailbox{}).Maybe()

	// 1) the generic two-arg Send(dest, msg)
	closureSend := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  process.options.Priority,
			Important: process.options.ImportantDelivery,
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
			Important: process.options.ImportantDelivery,
		}
		process.artifacts.Push(art)
	}

	// 3) SendImportant(dest, msg)
	closureSendImportant := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  process.options.Priority,
			Important: true,
		}
		process.artifacts.Push(art)
	}
	closureSendAfter := func(args mock.Arguments) {
		art := ArtifactSend{
			From:      pid,
			To:        args.Get(0),
			Message:   args.Get(1),
			Priority:  process.options.Priority,
			Important: process.options.ImportantDelivery,
			After:     args.Get(2).(time.Duration),
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
		Run(closureSendAfter).
		Return(gen.CancelFunc(nil), nil).
		Maybe()

	process.On("SendEvent", mock.AnythingOfType("gen.Atom"), mock.AnythingOfType("gen.Ref"), mock.Anything).
		Run(func(args mock.Arguments) {
			art := ArtifactEvent{
				Name:     args.Get(0).(gen.Atom),
				Message:  args.Get(2),
				Priority: process.options.Priority,
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
				Priority: process.options.Priority,
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
				Priority: process.options.Priority,
			}
			process.artifacts.Push(art)
		}).
		Return(nil).
		Maybe()

	process.On("Call", mock.AnythingOfType("gen.PID"), mock.Anything).
		Return(func(to any, request any) (any, error) {
			art := ArtifactCall{
				From:    pid,
				To:      to,
				Request: request,
			}
			process.artifacts.Push(art)
			for _, helper := range process.options.Helpers.Call {
				if eq := reflect.DeepEqual(art.Request, helper.Request); eq == false {
					continue
				}
				return helper.Response, nil
			}

			return nil, nil
		}).
		Maybe()

	process.On("Call", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Return(func(to any, request any) (any, error) {
			art := ArtifactCall{
				From:    pid,
				To:      to,
				Request: request,
			}
			process.artifacts.Push(art)
			for _, helper := range process.options.Helpers.Call {
				if eq := reflect.DeepEqual(art.Request, helper.Request); eq == false {
					continue
				}
				return helper.Response, nil
			}

			return nil, nil
		}).
		Maybe()

	closureCallTimeout := func(to any, request any, _ int) (any, error) {
		art := ArtifactCall{
			From:    pid,
			To:      to,
			Request: request,
		}
		process.artifacts.Push(art)
		for _, helper := range process.options.Helpers.Call {
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
	process.On("Call", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Return(closureCall).
		Maybe()
	process.On("Call", mock.AnythingOfType("gen.Alias"), mock.Anything).Return(closureCall).Maybe()
	process.
		On("CallPID", mock.AnythingOfType("gen.PID"), mock.Anything, mock.AnythingOfType("int")).
		Return(closureCallPID).Maybe()
	process.
		On(
			"CallProcessID",
			mock.AnythingOfType("gen.ProcessID"),
			mock.Anything,
			mock.AnythingOfType("int"),
		).
		Return(closureCallProcessID).Maybe()
	process.
		On(
			"CallAlias",
			mock.AnythingOfType("gen.Alias"),
			mock.Anything,
			mock.AnythingOfType("int"),
		).
		Return(closureCallAlias).Maybe()

	// monitor

	closureMonitor := func(args mock.Arguments) {
		art := ArtifactMonitor{
			Target: args.Get(0),
		}
		process.artifacts.Push(art)
	}
	closureDemonitor := func(args mock.Arguments) {
		art := ArtifactDemonitor{
			Target: args.Get(0),
		}
		process.artifacts.Push(art)
	}

	process.On("Monitor", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("Monitor", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("Monitor", mock.AnythingOfType("gen.Atom"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("Monitor", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()

	process.On("Demonitor", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("Demonitor", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("Demonitor", mock.AnythingOfType("gen.Atom"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("Demonitor", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()

	process.On("MonitorPID", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("MonitorProcessID", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("MonitorAlias", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureMonitor).Return(nil).Maybe()
	process.On("MonitorEvent", mock.AnythingOfType("gen.Event")).
		Run(closureMonitor).Return([]gen.MessageEvent{}, nil).Maybe()
	process.On("MonitorNode", mock.AnythingOfType("gen.Atom")).
		Run(closureMonitor).Return(nil).Maybe()

	process.On("DemonitorPID", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("DemonitorProcessID", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("DemonitorAlias", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("DemonitorEvent", mock.AnythingOfType("gen.Event")).
		Run(closureDemonitor).Return(nil).Maybe()
	process.On("DemonitorNode", mock.AnythingOfType("gen.Atom")).
		Run(closureDemonitor).Return(nil).Maybe()

	// Link

	closureLink := func(args mock.Arguments) {
		art := ArtifactLink{
			Target: args.Get(0),
		}
		process.artifacts.Push(art)
	}
	closureUnlink := func(args mock.Arguments) {
		art := ArtifactLink{
			Target: args.Get(0),
		}
		process.artifacts.Push(art)
	}

	process.On("Link", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("Link", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("Link", mock.AnythingOfType("gen.Atom"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("Link", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()

	process.On("Unlink", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("Unlink", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("Unlink", mock.AnythingOfType("gen.Atom"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("Unlink", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()

	process.On("LinkPID", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("LinkProcessID", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("LinkAlias", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureLink).Return(nil).Maybe()
	process.On("LinkEvent", mock.AnythingOfType("gen.Event")).
		Run(closureLink).Return([]gen.MessageEvent{}, nil).Maybe()
	process.On("LinkNode", mock.AnythingOfType("gen.Atom")).
		Run(closureLink).Return(nil).Maybe()

	process.On("UnlinkPID", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("UnlinkProcessID", mock.AnythingOfType("gen.ProcessID"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("UnlinkAlias", mock.AnythingOfType("gen.Alias"), mock.Anything).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("UnlinkEvent", mock.AnythingOfType("gen.Event")).
		Run(closureUnlink).Return(nil).Maybe()
	process.On("UnlinkNode", mock.AnythingOfType("gen.Atom")).
		Run(closureUnlink).Return(nil).Maybe()

	return process
}

// TODO:
// Forward(to PID, message *MailboxMessage, priority MessagePriority) error
// MetaInfo(meta Alias) (MetaInfo, error)
// Info() (ProcessInfo, error)
// RegisterEvent(name Atom, options EventOptions) (Ref, error)
// UnregisterEvent(name Atom) error
// Inspect(target PID, item ...string) (map[string]string, error)
// InspectMeta(meta Alias, item ...string) (map[string]string, error)
// Events() []Atom
// Aliases() []Alias
// DeleteAlias(alias Alias) error
// CreateAlias() (Alias, error)
// EnvList() map[Env]any
// RemoteSpawnRegister(node Atom, name Atom, register Atom, options ProcessOptions, args ...any) (PID, error)
// RemoteSpawn(node Atom, name Atom, options ProcessOptions, args ...any) (PID, error)
// SpawnMeta(behavior MetaBehavior, options MetaOptions) (Alias, error)
// SpawnRegister(register Atom, factory ProcessFactory, options ProcessOptions, args ...any) (PID, error)
