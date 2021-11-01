package gcp

import (
	"fmt"

	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

type cloudRunDependency struct {
	baseDir string
	Service string `yaml:"name"`
	EnvVar  string `yaml:"env"`
}

func (rt *cloudRunDependency) Load(d *api.ResourceDefinition) ([]api.Resource, []api.DependencyBinding, error) {
	services := []*cloudRunDependency{}
	var bindings []api.DependencyBinding
	err := yaml.Unmarshal(d.ServiceConfig, &services)
	if err != nil {
		return nil, nil, err
	}
	for _, service := range services {
		service.baseDir = rt.baseDir
		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   api.ReadWrite,
			Identity:     api.ResourceIdentity{Type: "cloudrun", ID: service.Service},
			Config:       service,
		})
	}

	return []api.Resource{}, bindings, nil
}

func (rt *cloudRunDependency) Name() string {
	return "cloudrun"
}

func (rt *cloudRunDependency) ConfigureResource(resource api.Resource) error {
	cloudRun, ok := resource.(*cloudRunConfig)
	if ok {
		serviceKey := rt.Service
		if rt.EnvVar != "" {
			serviceKey = rt.EnvVar
		}
		dependsOnLink := fmt.Sprintf("module.%s-%s", "cloudrun", rt.Service)
		urlLink := fmt.Sprintf("module.%s-%s.cloud_run_endpoint", "cloudrun", rt.Service)
		cloudRun.DependsOn = append(cloudRun.DependsOn, dependsOnLink)
		cloudRun.Env.Refs[fmt.Sprintf("%s_HOST", serviceKey)] = urlLink
	}
	return nil
}
