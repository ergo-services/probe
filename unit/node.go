package unit

import (
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/probe/unit/stub"
	"testing"
	"time"
)

type node struct {
	*stub.Node
}

func newNode(t testing.TB, artifacts lib.QueueMPSC) *node {
	nodeName := gen.Atom("unit-node@localhost")
	creation := time.Now().Unix()

	stubNode := &node{
		Node: stub.NewNode(t),
	}

	stubNodeLog := newStubLog(t, artifacts)
	stubNodeLog.SetLevel(gen.LogLevelTrace)
	stubNode.On("Log").Return(stubNodeLog).Maybe()
	stubNode.On("Name").Return(nodeName).Maybe()
	stubNode.On("Creation").Return(creation).Maybe()
	return stubNode
}
