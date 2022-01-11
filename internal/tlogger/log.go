package tlogger

import (
	"os"
	"runtime/debug"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Log is the default logger for apps
var Log log.Logger

var hlog log.Logger

// applyLogLevel applies min logging level
var ApplyLogLevel func(string)

func init() {
	Log = log.NewLogfmtLogger(os.Stdout)
	hlog = log.With(Log, "ts", log.DefaultTimestampUTC, "caller", log.Caller(6))
	Log = log.With(Log, "ts", log.DefaultTimestampUTC, "caller", log.Caller(5))

	ApplyLogLevel = func(lvl string) {
		switch lvl {
		case "debug":
			Log = level.NewFilter(Log, level.AllowDebug())
			hlog = level.NewFilter(hlog, level.AllowDebug())
		case "warn":
			Log = level.NewFilter(Log, level.AllowWarn())
			hlog = level.NewFilter(hlog, level.AllowWarn())
		case "error":
			Log = level.NewFilter(Log, level.AllowError())
			hlog = level.NewFilter(hlog, level.AllowError())
		case "all":
			Log = level.NewFilter(Log, level.AllowAll())
			hlog = level.NewFilter(hlog, level.AllowAll())
		default:
			Log = level.NewFilter(Log, level.AllowInfo())
			hlog = level.NewFilter(hlog, level.AllowInfo())
		}
		ApplyLogLevel = func(string) {}
	}
}

// Debug add a log entry w/ Debug level
func Debug(keyvals ...interface{}) {
	level.Debug(hlog).Log(keyvals...)
}

// Info add a log entry w/ Info level
func Info(keyvals ...interface{}) {
	level.Info(hlog).Log(keyvals...)
}

// Warn add a log entry w/ Warn level
func Warn(keyvals ...interface{}) {
	level.Warn(hlog).Log(keyvals...)
}

// Error add a log entry w/ Error level
func Error(keyvals ...interface{}) {
	level.Error(hlog).Log(keyvals...)
}

// Fatal add a log entry w/ Error level and exits
func Fatal(keyvals ...interface{}) {
	debug.PrintStack()
	level.Error(hlog).Log(keyvals...)
	os.Exit(1)
}

// FatalIf prints a fatal Error level and exits if err != nil
func FatalIf(err error) {
	if err == nil {
		return
	}
	level.Error(hlog).Log("err", err)
	os.Exit(1)
}
