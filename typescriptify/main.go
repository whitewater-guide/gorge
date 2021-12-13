package main

import (
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
	"github.com/whitewater-guide/gorge/core"
)

func main() {
	converter := typescriptify.New()
	converter.Add(core.Gauge{})
	converter.Add(core.Measurement{})
	converter.Add(core.JobDescription{})
	converter.Add(core.UnhealthyJob{})
	converter.Add(core.ScriptDescriptor{})
	converter.Add(core.Status{})
	converter.Add(core.ErrorResponse{})
	converter.CreateInterface = true
	err := converter.ConvertToFile("index.d.ts")
	if err != nil {
		panic(err.Error())
	}
}
