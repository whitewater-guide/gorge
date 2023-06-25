//go:build tools
// +build tools

package tools

import (
	_ "github.com/cortesi/modd"
	_ "github.com/cortesi/modd/cmd/modd"
	_ "github.com/evanoberholster/timezoneLookup/v2/cmd"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/oligot/go-mod-upgrade"
)
