package unit

import (
	"fmt"
	"testing"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
	"ergo.services/testing/unit/stub"
	"github.com/stretchr/testify/mock"
)

type log struct {
	*stub.Log
	level gen.LogLevel
}

func newStubLog(t testing.TB, artifacts lib.QueueMPSC) *log {
	stubLog := &log{
		Log: stub.NewLog(t),
	}

	stubLog.
		On("SetLevel", mock.AnythingOfType("gen.LogLevel")).
		Return(nil).
		Run(func(args mock.Arguments) {
			stubLog.level = args.Get(0).(gen.LogLevel)
		}).
		Maybe()

	stubLog.
		On("Level").
		Return(func() gen.LogLevel {
			return stubLog.level
		}).Maybe()

	stubLog.
		On("Trace", mock.IsType("string"), mock.Anything).
		Return(mock.Anything).
		Maybe()

	stubLog.
		On("Info", mock.IsType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			if stubLog.level > gen.LogLevelInfo {
				return
			}
			format := args.Get(0).(string)
			params := args.Get(1).([]any)
			art := ArtifactLog{
				Level:   gen.LogLevelInfo,
				Message: fmt.Sprintf(format, params...),
			}
			artifacts.Push(art)
		}).
		Maybe()

	stubLog.
		On("Debug", mock.IsType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			if stubLog.level > gen.LogLevelInfo {
				return
			}
			format := args.Get(0).(string)
			params := args.Get(1).([]any)
			art := ArtifactLog{
				Level:   gen.LogLevelDebug,
				Message: fmt.Sprintf(format, params...),
			}
			artifacts.Push(art)
		}).
		Maybe()

	stubLog.
		On("Warning", mock.IsType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			if stubLog.level > gen.LogLevelInfo {
				return
			}
			format := args.Get(0).(string)
			params := args.Get(1).([]any)
			art := ArtifactLog{
				Level:   gen.LogLevelWarning,
				Message: fmt.Sprintf(format, params...),
			}
			artifacts.Push(art)
		}).
		Maybe()

	stubLog.
		On("Error", mock.IsType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			if stubLog.level > gen.LogLevelInfo {
				return
			}
			format := args.Get(0).(string)
			params := args.Get(1).([]any)
			art := ArtifactLog{
				Level:   gen.LogLevelError,
				Message: fmt.Sprintf(format, params...),
			}
			artifacts.Push(art)
		}).
		Maybe()

	stubLog.
		On("Panic", mock.IsType("string"), mock.Anything).
		Run(func(args mock.Arguments) {
			if stubLog.level > gen.LogLevelInfo {
				return
			}
			format := args.Get(0).(string)
			params := args.Get(1).([]any)
			art := ArtifactLog{
				Level:   gen.LogLevelPanic,
				Message: fmt.Sprintf(format, params...),
			}
			artifacts.Push(art)
		}).
		Maybe()

	return stubLog
}
