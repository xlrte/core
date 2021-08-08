package main

import (
	"fmt"

	"github.com/xlrte/core/pkg/api"
)

type runtime string

func (g runtime) Greet() {
	fmt.Println("Hello Universe")
}

func (g runtime) SupportServiceRuntime(name string) bool {
	return false
}
func (g runtime) Resources() []api.Resource {
	return []api.Resource{}
}
func (g runtime) Configure(name string, artifact api.Artifact, env api.Env, previousOutputs []api.Output) error {
	return nil
}

// Runtime is the exported interface
var Runtime runtime //nolint
