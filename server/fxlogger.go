package main

import (
	"strings"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx/fxevent"
)

// fxLogger is an Fx event logger that attempts to write human-readable
// mesasges to the console.
type fxLogger struct {
	*logrus.Entry
}

// LogEvent logs the given event to the provided logrus logger
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.WithFields(logrus.Fields{
			"event":    "OnStartExecuting",
			"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
			"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
		}).Debug("executing")
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.WithFields(logrus.Fields{
				"event":    "OnStartExecuted",
				"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
				"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
				"runtime":  e.Runtime,
			}).Errorf("failed: %v", e.Err)
		} else {
			l.WithFields(logrus.Fields{
				"event":    "OnStartExecuted",
				"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
				"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
				"runtime":  e.Runtime,
			}).Debug("ran successfully")
		}
	case *fxevent.OnStopExecuting:
		l.WithFields(logrus.Fields{
			"event":    "OnStopExecuting",
			"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
			"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
		}).Debug("executing")
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.WithFields(logrus.Fields{
				"event":    "OnStopExecuted",
				"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
				"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
				"runtime":  e.Runtime,
			}).Errorf("failed: %v", e.Err)
		} else {
			l.WithFields(logrus.Fields{
				"event":    "OnStopExecuted",
				"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
				"caller":   strings.Replace(e.CallerName, "github.com/whitewater-guide/gorge/", "", 1),
				"runtime":  e.Runtime,
			}).Debug("ran successfully")
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.WithFields(logrus.Fields{
				"event":    "Supplied",
				"typename": e.TypeName,
			}).Errorf("failed: %v", e.Err)
		} else {
			l.WithFields(logrus.Fields{
				"event":    "Supplied",
				"typename": e.TypeName,
			}).Debug("successfully")
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.WithField("event", "Provided").Debugf("%v <= %v", rtype, strings.Replace(e.ConstructorName, "github.com/whitewater-guide/gorge/", "", 1))
		}
		if e.Err != nil {
			l.WithField("event", "Provided").Errorf("Error after options were applied: %v", e.Err)
		}
	case *fxevent.Invoking:
		l.WithFields(logrus.Fields{
			"event":    "Invoking",
			"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
		}).Debug("invoking")
	case *fxevent.Invoked:
		if e.Err != nil {
			l.WithFields(logrus.Fields{
				"event":    "Invoked",
				"function": strings.Replace(e.FunctionName, "github.com/whitewater-guide/gorge/", "", 1),
				"trace":    e.Trace,
			}).Errorf("failed: %v", e.Err)
		}
	case *fxevent.Stopping:
		l.WithField("event", "Stopping").Debugf("%v", strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.WithField("event", "Stopping").Errorf("failed to stop cleanly: %v", e.Err)
		}
	case *fxevent.RollingBack:
		l.WithField("event", "RollingBack").Debugf("Start failed, rolling back: %v", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.WithField("event", "RollingBack").Errorf("Couldn't roll back cleanly: %v", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.WithField("event", "Started").Errorf("failed to start: %v", e.Err)
		} else {
			l.WithField("event", "Started").Debug("running")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.WithField("event", "LoggerInitialized").Errorf("Failed to initialize custom logger: %+v", e.Err)
		} else {
			l.WithFields(logrus.Fields{
				"event":       "LoggerInitialized",
				"constructor": e.ConstructorName,
			}).Debugf("initialized custom logger")
		}
	}
}

func newFxLogger(l *logrus.Logger) fxevent.Logger {
	return &fxLogger{l.WithField("logger", "fx")}
}
