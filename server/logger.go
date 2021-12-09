package main

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/config"
)

func newLogger(cfg *config.Config) *logrus.Logger {
	result := logrus.New()
	if cfg.Log.Format == "json" {
		result.SetFormatter(&logrus.JSONFormatter{})
	} else {
		result.SetFormatter(&logrus.TextFormatter{ForceColors: true, DisableTimestamp: true})
	}
	logLevel := cfg.Log.Level
	if cfg.Debug {
		logLevel = "debug"
	}
	if logLevel == "" {
		result.SetOutput(ioutil.Discard)
	} else {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			lvl = logrus.DebugLevel
		}
		result.SetLevel(lvl)
	}
	return result
}

func testLogger(cfg *config.Config) *logrus.Logger {
	result := logrus.New()
	result.SetOutput(ioutil.Discard)
	return result
}
