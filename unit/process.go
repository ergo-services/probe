package unit

import (
	"ergo.services/ergo/gen"
	"ergo.services/probe/unit/stub"
)

type process struct {
	stub.Process

	pid      gen.PID
	mailbox  gen.ProcessMailbox
	behavior gen.ProcessBehavior
}
