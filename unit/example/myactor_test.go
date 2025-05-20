package main

import (
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/testing/unit"
)

func TestMyActor_Init(t *testing.T) {
	options := unit.SpawnOptions{
		Priority: gen.MessagePriorityNormal,
		CallHelpers: []unit.CallHelper{
			{Request: 12345, Response: 6789},
		},
	}
	process, err := unit.Spawn(t, factoryMyActor, options)
	if err != nil {
		t.Fatalf("unable to spawn: %s", err)
	}

	expected := []any{
		unit.ArtifactSend{From: process.PID(), To: process.PID(), Message: "hello"},
		unit.ArtifactLog{Level: gen.LogLevelDebug, Message: "actor started " + process.PID().String() + " 1"},
	}
	process.ValidateArtifacts(t, expected)

	behavior := process.Behavior().(*myActor)
	if behavior.value != 1 {
		t.Fatal("incorrect value")
	}

	t.Logf("started process: %s", process.PID())

	process.Log().SetLevel(gen.LogLevelWarning)
	behavior.value = 8
	behavior.HandleMessage(gen.PID{}, "increase")
	if behavior.value != 16 {
		t.Fatalf("hasn't been increased")
	}

	expected = []any{
		// unit.ArtifactLog{Level: gen.LogLevelDebug, Message: "actor started " + process.PID().String() + " 1"},
		unit.ArtifactSend{From: process.PID(), To: gen.Atom("abc"), Message: behavior.value},
		unit.ArtifactCall{From: process.PID(), To: gen.ProcessID{}, Request: 12345},
	}
	process.ValidateArtifacts(t, expected)

}
