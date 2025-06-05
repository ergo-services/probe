package unit

import (
	"time"

	"ergo.services/ergo/gen"
)

type ArtifactSend struct {
	From      gen.PID
	To        any
	Message   any
	Priority  gen.MessagePriority // SendWithPriority
	Important bool                // SendImportant
	Ref       gen.Ref             // SendResponse
	After     time.Duration       // SendAfter
}

type ArtifactSpawn struct {
	Factory gen.ProcessFactory
	Options gen.ProcessOptions
	Args    []any
}

type ArtifactSpawnMeta struct {
	Factory gen.MetaBehavior
	Options gen.MetaOptions
}

type ArtifactLog struct {
	Level   gen.LogLevel
	Message string
}

type ArtifactEvent struct {
	Name     gen.Atom
	Message  any
	Priority gen.MessagePriority
}

type ArtifactExit struct {
	To     gen.PID
	Reason error
}

type ArtifactExitMeta struct {
	Meta   gen.Alias
	Reason error
}

type ArtifactCall struct {
	From     gen.PID
	To       any
	Request  any
	Priority gen.MessagePriority
}

type ArtifactMonitor struct {
	Target any
}

type ArtifactDemonitor struct {
	Target any
}

type ArtifactLink struct {
	Target any
}

type ArtifactUnink struct {
	Target any
}
