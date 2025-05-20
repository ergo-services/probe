package unit

import (
	"ergo.services/ergo/gen"
)

type ArtifactSend struct {
	From    gen.PID
	To      any
	Message any
}

type ArtifactLog struct {
	Level   gen.LogLevel
	Message string
}

type ArtifactCall struct {
	From     gen.PID
	To       any
	Request  any
	Priority gen.MessagePriority
}
