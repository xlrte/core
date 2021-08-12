package api

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

func ParseDeploymentConfig(rootDir string, selector TriggerSelector, runtimes *Runtimes) ([]*DeploymentConfig, error) {
	envs, err := ReadAllEnvironments(filepath.Join(rootDir, "environments"))
	if err != nil {
		return nil, err
	}
	svcs, err := ReadAllServices(filepath.Join(rootDir, "services"))
	if err != nil {
		return nil, err
	}
	var targetEnv *Environment
	for _, env := range envs {
		if selector.IsMatch(env) {
			targetEnv = &env
		}
	}
	if targetEnv == nil {
		return nil, fmt.Errorf("could not find a target environment for %v", selector)
	}
	if len(svcs) == 0 {
		return nil, fmt.Errorf("no services to deploy")
	}

	runtimeMap := make(map[string]*DeploymentConfig)
	for _, svc := range svcs {
		rtme, err := runtimes.getRuntimeFor(&svc)
		if err != nil {
			return nil, err
		}
		conf := runtimeMap[rtme.Name()]
		if conf != nil {
			conf = &DeploymentConfig{
				Runtime:     rtme,
				Services:    []*Service{&svc},
				Environment: targetEnv,
			}
			runtimeMap[rtme.Name()] = conf
		} else {
			conf.Services = append(conf.Services, &svc)
		}

		for k, v := range svc.DependsOn {
			_, err := supportsResourceType(rtme, svc.Name(), k)
			if err != nil {
				return nil, err
			}
			bytes, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			inf := targetEnv.Resources[k]
			if inf != nil {
				resourceBytes, err := json.Marshal(inf)
				if err != nil {
					return nil, err
				}
				resource := ResourceDefinition{serviceConfig: bytes, resourceConfig: &resourceBytes, name: k}
				conf.Resources = append(conf.Resources, &resource)
			} else {
				resource := ResourceDefinition{serviceConfig: bytes, resourceConfig: nil, name: k}
				conf.Resources = append(conf.Resources, &resource)
			}
		}
	}

	configs := []*DeploymentConfig{}
	for _, v := range runtimeMap {
		configs = append(configs, v)
	}
	return configs, nil
}

func supportsResourceType(runtime Runtime, service, name string) (ResourceLoader, error) {
	for _, resource := range runtime.Resources() {
		if resource.Name() == name {
			return resource, nil
		}
	}
	return nil, fmt.Errorf("runtime %v that matches service %v does not support resource-type %v. resources attached to a service must belong to the same runtime", runtime.Name(), service, name)

}

func (runtimes *Runtimes) getRuntimeFor(service *Service) (Runtime, error) {
	for _, rtime := range runtimes.Runtimes {
		if rtime.SupportServiceRuntime(service.Runtime) {
			return rtime, nil
		}
	}
	return nil, fmt.Errorf("could not find a Runtime that supports service of type %s", service.Runtime)
}
