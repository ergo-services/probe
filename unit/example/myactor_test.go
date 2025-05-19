package main

import (
	"ergo.services/ergo/gen"
	"ergo.services/probe/unit"
	"testing"
)

func TestMActor_Init(t *testing.T) {
	node := unit.StartNode(t, "node-sub@localhost", gen.NodeOptions{})

	pid, proc := node.Spawn(factoryMyActor) // ProcessInit + Init already run

	node.HasProcessInit(t, pid) // ← NEW single‑call assertion

	// still check logging if you like
	node.MockLog(proc).AssertNumberOfCalls(t, "Info", 1)
	proc.AssertExpectations(t)
}
