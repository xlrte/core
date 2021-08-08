package gcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/xlrte/core/pkg/api"
)

type CloudRun struct {
}
type CloudRunResources struct {
}

type runtime string

func (g runtime) Greet() {
	fmt.Println("Hello Universe")
}

func (g runtime) SupportServiceRuntime(name string) bool {
	return name == "cloudrun"
}

func (g runtime) Resources() []api.Resource {
	return []api.Resource{}
}

func (g runtime) Configure(name string, artifact api.Artifact, env api.Env, previousOutputs []api.Output) error {
	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	return nil
}

// Runtime is the exported interface
var Runtime runtime //nolint
