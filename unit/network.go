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

// TODO:
// Registrar() (Registrar, error)
// Cookie() string
// SetCookie(cookie string) error
// MaxMessageSize() int
// SetMaxMessageSize(size int)
// NetworkFlags() NetworkFlags
// SetNetworkFlags(flags NetworkFlags)
// Acceptors() ([]Acceptor, error)
// Node(name Atom) (RemoteNode, error)
// GetNode(name Atom) (RemoteNode, error)
// GetNodeWithRoute(name Atom, route NetworkRoute) (RemoteNode, error)
// Nodes() []Atom
// AddRoute(match string, route NetworkRoute, weight int) error
// RemoveRoute(match string) error
// Route(name Atom) ([]NetworkRoute, error)
// AddProxyRoute(match string, proxy NetworkProxyRoute, weight int) error
// RemoveProxyRoute(match string) error
// ProxyRoute(name Atom) ([]NetworkProxyRoute, error)
// RegisterProto(proto NetworkProto)
// RegisterHandshake(handshake NetworkHandshake)
// EnableSpawn(name Atom, factory ProcessFactory, nodes ...Atom) error
// DisableSpawn(name Atom, nodes ...Atom) error
// EnableApplicationStart(name Atom, nodes ...Atom) error
// DisableApplicationStart(name Atom, nodes ...Atom) error
// Info() (NetworkInfo, error)
// Mode() NetworkMode
