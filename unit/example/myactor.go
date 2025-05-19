package main

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factoryMyActor() gen.ProcessBehavior {
	return &myActor{}
}

type myActor struct {
	act.Actor
}
