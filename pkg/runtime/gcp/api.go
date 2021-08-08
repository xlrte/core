package gcp

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/xlrte/core/pkg/api"
)

type CloudRun struct {
}
type CloudRunResources struct {
}

type GcpRuntime struct {
}

func (g *GcpRuntime) SupportServiceRuntime(name string) bool {
	return name == "cloudrun"
}

func (g *GcpRuntime) Resources() []api.Resource {
	return []api.Resource{}
}

func (g *GcpRuntime) Configure(name string, artifact api.Artifact, env api.EnvVars) error {
	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	return nil
}
