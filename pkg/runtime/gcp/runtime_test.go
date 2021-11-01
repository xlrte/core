package gcp

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

func Test_toCloudRunSettings(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-config.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	rte := cloudRunLoader{baseDir: "."}

	conf, err := rte.toCloudRunSettings(api.EnvContext{Version: func(s string) (string, error) { return "v1", nil }, EnvName: "prod", Context: "theproject", Region: "europe-west6"},
		&api.Service{
			SVCName: "cloudrun-srv",
			Runtime: "cloudrun",
			Spec: cloudRunSpec{
				BaseName: "foo",
			},
			DependsOn: make(map[string]interface{}),
			Env:       api.EnvVars{},
		}, api.DeploymentContext{Env: api.EnvVars{}, Resources: &bytes})
	assert.NoError(t, err)
	assert.Equal(t, "cloudrun-srv", conf.ServiceName)
	assert.Equal(t, "gcr.io/theproject/foo:v1", conf.ImageID)
	assert.Equal(t, 100, conf.Traffic)
	assert.False(t, conf.IsPublic)

	assert.Equal(t, 2, conf.RuntimeConfig.CPU)
	assert.Equal(t, "1024Mi", conf.RuntimeConfig.Memory)
	assert.Equal(t, 10, conf.RuntimeConfig.MaxRequests)
	assert.Equal(t, 100, conf.RuntimeConfig.Timeout)
	assert.Equal(t, 10, conf.RuntimeConfig.Scaling.MaxInstances)
	assert.Equal(t, 1, conf.RuntimeConfig.Scaling.MinInstances)
	assert.Equal(t, "cloudrun-srv", conf.RuntimeConfig.Name)

}

func Test_toCloudRunSettings_No_CPU(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-nocpu.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	rte := cloudRunLoader{baseDir: "."}

	service := &api.Service{
		SVCName: "cloudrun-srv",
		Runtime: "cloudrun",
		Spec: cloudRunSpec{
			BaseName: "gcr.io/xlrte/foo",
			Http:     http{Public: true},
		},
		DependsOn: make(map[string]interface{}),
		Env:       api.EnvVars{},
	}

	conf, err := rte.toCloudRunSettings(
		api.EnvContext{Version: func(s string) (string, error) { return "v1", nil }, EnvName: "prod", Context: "theproject", Region: "europe-west6"},
		service, api.DeploymentContext{Env: api.EnvVars{}, Resources: &bytes})
	assert.NoError(t, err)
	assert.Equal(t, 1, conf.RuntimeConfig.CPU)
	assert.Equal(t, "512Mi", conf.RuntimeConfig.Memory)
	assert.Equal(t, 80, conf.RuntimeConfig.MaxRequests)
	assert.Equal(t, 300, conf.RuntimeConfig.Timeout)
	assert.Equal(t, 1000, conf.RuntimeConfig.Scaling.MaxInstances)
	assert.Equal(t, 0, conf.RuntimeConfig.Scaling.MinInstances)
	assert.Equal(t, "cloudrun-srv", conf.RuntimeConfig.Name)
	assert.True(t, conf.IsPublic)
}

func Test_Basics(t *testing.T) {
	rte := NewRuntime(".", ".")

	rt := rte.(*gcpRuntime)
	assert.NotNil(t, rt)

	assert.Len(t, rt.Services(), 1)
	assert.Equal(t, rt.Name(), "gcp")
}

func Test_CreateParent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()
	rte := NewRuntime(tmpDir, tmpDir)

	env := api.EnvContext{
		Context: "cloudrun-test",
		Region:  "europe-west6",
		EnvName: "prod",
		Version: func(s string) (string, error) { return "v1", nil },
	}
	err = rte.Init(env)
	assert.NoError(t, err)

	service := &api.Service{
		SVCName: "cloudrun-srv",
		Runtime: "cloudrun",
		Spec: cloudRunSpec{
			BaseName: "foo",
			Http:     http{Public: true},
		},
		DependsOn: make(map[string]interface{}),
		Env:       api.EnvVars{},
	}

	rt := &cloudRunLoader{tmpDir, service}

	resource, err := rt.Load(env, service, api.DeploymentContext{Env: api.EnvVars{}})

	assert.NoError(t, err)
	err = resource.Configure()
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(tmpDir, "main.tf"))
	assert.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "main.tf")) // nolint
	assert.NoError(t, err)
	str := string(b)
	assert.Contains(t, str, "cloudrun-cloudrun-srv")

}

func Test_ConfigureService_With_Dependencies(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-nocpu.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	rte := cloudRunLoader{baseDir: tmpDir}

	service := &api.Service{
		SVCName: "cloudrun-srv",
		Runtime: "cloudrun",
		Spec: cloudRunSpec{
			BaseName: "gcr.io/xlrte/foo",
			Http:     http{Public: true},
		},
		DependsOn: make(map[string]interface{}),
		Env:       api.EnvVars{},
	}

	service2 := &api.Service{
		SVCName: "cloudrun-srv2",
		Runtime: "cloudrun",
		Spec: cloudRunSpec{
			BaseName: "gcr.io/xlrte/foo",
			Http:     http{Public: true},
		},
		DependsOn: make(map[string]interface{}),
		Env:       api.EnvVars{},
	}

	_, err = rte.Load(
		api.EnvContext{Version: func(s string) (string, error) { return "v1", nil }, EnvName: "prod", Context: "theproject", Region: "europe-west6"},
		service, api.DeploymentContext{Env: api.EnvVars{}, Resources: &bytes})
	assert.NoError(t, err)
	_, err = rte.Load(
		api.EnvContext{Version: func(s string) (string, error) { return "v1", nil }, EnvName: "prod", Context: "theproject", Region: "europe-west6"},
		service2, api.DeploymentContext{Env: api.EnvVars{}, Resources: &bytes},
	)
	assert.NoError(t, err)

}

func Test_Network_Configuration(t *testing.T) {
	resourceData, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "resources.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(resourceData, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	rte := cloudRunLoader{baseDir: "."}

	service := &api.Service{
		SVCName: "cloudrun-srv",
		Runtime: "cloudrun",
		Spec: cloudRunSpec{
			BaseName: "gcr.io/xlrte/foo",
			Http:     http{Public: true},
		},
		DependsOn: make(map[string]interface{}),
		Env:       api.EnvVars{},
	}

	conf, err := rte.toCloudRunSettings(
		api.EnvContext{Version: func(s string) (string, error) { return "v1", nil }, EnvName: "prod", Context: "theproject", Region: "europe-west6"},
		service, api.DeploymentContext{Env: api.EnvVars{}, Resources: &bytes})
	assert.NoError(t, err)
	assert.Equal(t, conf.NetworkConfig.DNSName, "xlrte.org.")
	assert.Equal(t, conf.NetworkConfig.Domain.DNSZone, "xlrte_zone")
	assert.Equal(t, conf.NetworkConfig.Domain.Name, "xlrte.org")

}
