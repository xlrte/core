package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var selector = NameSelector{
	EnvName: "prod",
}

type dummyRuntime struct {
}

func Test_No_Matching_Env_Found(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{}},
	}

	_, err := parseDeploymentConfig("testdata", &NameSelector{"foo"}, &runtimes)
	assert.Error(t, err)
}

func Test_No_Services_Found(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{}},
	}

	_, err := parseDeploymentConfig(filepath.Join("testdata", "noservices"), &selector, &runtimes)
	assert.Error(t, err)
}

func (rt *dummyRuntime) Name() string {
	return "dummy"
}

func (rt *dummyRuntime) SupportServiceRuntime(name string) bool {
	return true
}
func (rt *dummyRuntime) Resources() []ResourceLoader {
	return nil
}
func (rt *dummyRuntime) Configure(name string, artifact Artifact, env EnvVars) error {
	return nil
}
