package unit

import (
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"sync/atomic"
	"testing"
	"time"
)

type node struct {
	*stub.Node
	uniqID uint64
}

func newNode(t testing.TB, artifacts lib.QueueMPSC) *node {
	nodeName := gen.Atom("unit-node@localhost")
	creation := time.Now().Unix()

	stubNode := &node{
		Node: stub.NewNode(t),
	}

	virtualPID := gen.PID{Node: nodeName, ID: 1, Creation: creation}
	stubNodeLog := newStubLog(t, artifacts)
	stubNodeLog.SetLevel(gen.LogLevelTrace)
	stubNode.On("Log").Return(stubNodeLog).Maybe()
	stubNode.On("Name").Return(nodeName).Maybe()
	stubNode.On("Creation").Return(creation).Maybe()
	stubNode.On("PID").Return(virtualPID).Maybe()
	stubNode.On("MakeRef").Return(func() gen.Ref {
		// copy-pasted from the real MakeRef
		var ref gen.Ref
		ref.Node = nodeName
		ref.Creation = creation
		id := atomic.AddUint64(&stubNode.uniqID, 1)
		ref.ID[0] = id & ((2 << 17) - 1)
		ref.ID[1] = id >> 46
		return ref
	}).Maybe()

	stubNetwork := newNetwork(t, artifacts)
	stubNode.On("Network").Return(stubNetwork).Maybe()
	return stubNode
}
