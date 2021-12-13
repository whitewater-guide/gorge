package config

import (
	"go.uber.org/fx"
)

var TestModule = fx.Provide(TestConfig)
