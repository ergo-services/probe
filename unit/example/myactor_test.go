package main

import (
	"ergo.services/ergo/gen"
	"ergo.services/probe/unit"
	"testing"
)

func TestMyActor_Init(t *testing.T) {
	options := unit.SpawnOptions{}
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
	}
	process.ValidateArtifacts(t, expected)

}
