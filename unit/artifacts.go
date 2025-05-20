package unit

import (
	"ergo.services/ergo/gen"
)

type ArtifactSend struct {
	From      gen.PID
	To        any
	Message   any
	Priority  gen.MessagePriority // SendWithPriority
	Important bool                // SendImportant
	Ref       gen.Ref             // SendResponse
}

type ArtifactLog struct {
	Level   gen.LogLevel
	Message string
}

type ArtifactEvent struct {
	Name    gen.Atom
	Token   gen.Ref
	Message any
}

type ArtifactExit struct {
	To     gen.PID
	Reason error
}

type ArtifactExitMeta struct {
	Meta   gen.Alias
	Reason error
}
