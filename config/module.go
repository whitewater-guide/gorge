package config

import "go.uber.org/fx"

var Module = fx.Provide(newConfig)
var TestModule = fx.Provide(testConfig)
