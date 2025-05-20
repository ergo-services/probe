package unit

import (
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"testing"
)

type network struct {
	*stub.Network
}

func newNetwork(t testing.TB, artifacts lib.QueueMPSC) *network {
	stubNetwork := &network{
		Network: stub.NewNetwork(t),
	}

	return stubNetwork
}
