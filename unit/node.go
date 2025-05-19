package unit

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/probe/unit/stub"
	"github.com/stretchr/testify/mock"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func StartNode(t testing.TB, name gen.Atom, options gen.NodeOptions) *Node {
	nm := stub.NewNode(t)
	nm.On("Name").Return(gen.Atom(name)).Maybe()

	return &Node{
		t:     t,
		name:  gen.Atom(name),
		node:  nm,
		mail:  make(map[gen.PID]gen.ProcessMailbox),
		procs: make(map[gen.PID]*stub.Process),
		behs:  make(map[gen.PID]act.ActorBehavior),
	}
}

type Node struct {
	t      testing.TB
	name   gen.Atom
	nextID uint64

	node *stub.Node

	mail  map[gen.PID]gen.ProcessMailbox // PID → mailbox
	procs map[gen.PID]*stub.Process      // PID → mock process
	behs  map[gen.PID]act.ActorBehavior  // PID → concrete behaviour (if actor)
}

func (n *Node) Spawn(factory func() gen.ProcessBehavior) (gen.PID, *stub.Process) {
	pid := gen.PID{Node: n.name, ID: atomic.AddUint64(&n.nextID, 1)}
	beh := factory()

	// — mailbox —
	mb := gen.ProcessMailbox{
		Main:   lib.NewQueueMPSC(),
		System: lib.NewQueueMPSC(),
		Urgent: lib.NewQueueMPSC(),
		Log:    lib.NewQueueMPSC(),
	}
	n.mail[pid] = mb

	// — stubs —
	proc := stub.NewProcess(n.t)
	node := n.node
	log := stub.NewLog(n.t)
	start := time.Now()

	proc.On("Behavior").Return(beh).Maybe()
	proc.On("Mailbox").Return(mb).Maybe()
	proc.On("Panic").Return(nil).Maybe()
	proc.On("Log").Return(log).Maybe()
	proc.On("Node").Return(node).Maybe()
	proc.On("PID").Return(pid).Maybe()

	node.On("Name").Return(n.name).Maybe()
	log.On("Info", mock.Anything, mock.Anything).Maybe()
	log.On("Error", mock.Anything, mock.Anything).Maybe()

	proc.On("Info").Return(gen.ProcessInfo{
		PID:      pid,
		Name:     gen.Atom(reflect.TypeOf(beh).Elem().Name()),
		Behavior: reflect.TypeOf(beh).Elem().String(),
		Uptime:   int64(time.Since(start).Seconds()),
	}, nil).Maybe()

	// Send stub → artefact capture
	proc.On("Send", mock.AnythingOfType("gen.PID"), mock.Anything).
		Run(func(args mock.Arguments) {
			dst := args.Get(0).(gen.PID)
			msg := args.Get(1)
			if dst == pid {
				mb.Main.Push(msg)
			} else {
				mb.Urgent.Push(struct {
					To  gen.PID
					Msg any
				}{dst, msg})
			}
		}).Return(nil).Maybe()

	// attach context
	if err := beh.ProcessInit(proc); err != nil {
		n.t.Fatalf("ProcessInit failed: %v", err)
	}

	// remember for later helpers
	n.procs[pid] = proc
	if ab, ok := beh.(act.ActorBehavior); ok {
		n.behs[pid] = ab
	}

	return pid, proc
}

func (n *Node) HasProcessInit(t testing.TB, pid gen.PID) {
	mb := n.mail[pid]
	if got := mb.Main.Len(); got != 1 {
		t.Fatalf("expected 1 self‑message, got %d", got)
	}
	if msg, _ := mb.Main.Pop(); msg != "init" {
		t.Fatalf("expected \"init\", got %#v", msg)
	}
}

func (n *Node) MockLog(proc *stub.Process) *stub.Log {
	return proc.Log().(*stub.Log)
}

func (n *Node) ExpectMonitor(proc *stub.Process, ev gen.Event) {
	proc.On("MonitorEvent", ev).Return([]gen.MessageEvent{}, nil).Once()
}
