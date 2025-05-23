package unit

import (
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"github.com/stretchr/testify/mock"
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
	stubNodeLog.SetLevel(gen.LogLevelDebug)
	stubNode.On("Log").Return(stubNodeLog).Maybe()
	stubNode.On("Name").Return(nodeName).Maybe()
	stubNode.On("Creation").Return(creation).Maybe()
	stubNode.On("PID").Return(virtualPID).Maybe()
	stubNode.On("Uptime").Return(0).Maybe()
	stubNode.On("Version").Return(gen.Version{}).Maybe()
	stubNode.On("FrameworkVersion").Return(gen.Version{}).Maybe()
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

	stubNode.On("Send").Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     virtualPID,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: gen.MessagePriorityNormal,
		}
		artifacts.Push(art)
	}).Return(nil).Maybe()

	stubNode.On("SendWithPriority").Run(func(args mock.Arguments) {
		art := ArtifactSend{
			From:     virtualPID,
			To:       args.Get(0),
			Message:  args.Get(1),
			Priority: args.Get(2).(gen.MessagePriority),
		}
		artifacts.Push(art)
	}).Return(nil).Maybe()

	stubNetwork := newNetwork(t, artifacts)
	stubNode.On("Network").Return(stubNetwork).Maybe()
	return stubNode
}

// TODO
// Info() (NodeInfo, error)
// EnvList() map[Env]any
// SetEnv(name Env, value any)
// Env(name Env) (any, bool)
// Spawn(factory ProcessFactory, options ProcessOptions, args ...any) (PID, error)
// SpawnRegister(register Atom, factory ProcessFactory, options ProcessOptions, args ...any) (PID, error)
// RegisterName(name Atom, pid PID) error
// UnregisterName(name Atom) (PID, error)
// MetaInfo(meta Alias) (MetaInfo, error)
// ProcessInfo(pid PID) (ProcessInfo, error)
// ProcessList() ([]PID, error)
// ProcessListShortInfo(start, limit int) ([]ProcessShortInfo, error)
// ProcessState(pid PID) (ProcessState, error)
// ApplicationLoad(app ApplicationBehavior, args ...any) (Atom, error)
// ApplicationInfo(name Atom) (ApplicationInfo, error)
// SendEvent(name Atom, token Ref, options MessageOptions, message any) error
// RegisterEvent(name Atom, options EventOptions) (Ref, error)
// UnregisterEvent(name Atom) error
// SendExit(pid PID, reason error) error
// LogLevelProcess(pid PID) (LogLevel, error)
// SetLogLevelProcess(pid PID, level LogLevel) error
// LogLevelMeta(meta Alias) (LogLevel, error)
// SetLogLevelMeta(meta Alias, level LogLevel) error
// Loggers() []string
// LoggerAddPID(pid PID, name string, filter ...LogLevel) error
// LoggerAdd(name string, logger LoggerBehavior, filter ...LogLevel) error
// LoggerLevels(name string) []LogLevel
// Commercial() []Version
// SetCTRLC(enable bool)
// ApplicationUnload(name Atom) error
// ApplicationStart(name Atom, options ApplicationOptions) error
// ApplicationStartTemporary(name Atom, options ApplicationOptions) error
// ApplicationStartTransient(name Atom, options ApplicationOptions) error
// ApplicationStartPermanent(name Atom, options ApplicationOptions) error
// ApplicationStop(name Atom) error
// ApplicationStopForce(name Atom) error
// ApplicationStopWithTimeout(name Atom, timeout time.Duration) error
// Applications() []Atom
// ApplicationsRunning() []Atom
// NetworkStart(options NetworkOptions) error
// NetworkStop() error
// Cron() Cron
// CertManager() CertManager
// Security() SecurityOptions
// Stop()
// StopForce()
// Wait()
// WaitWithTimeout(timeout time.Duration) error
// Kill(pid PID) error
