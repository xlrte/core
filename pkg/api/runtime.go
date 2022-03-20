package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xlrte/core/pkg/api/secrets"
	"gopkg.in/yaml.v2"
)

type preApplyFn = func(ctx context.Context) error
type Command int64

const (
	Plan Command = iota
	Export
	Apply
	Delete
)

func InitEnvironment(ctx context.Context, baseDir, env, context, region string, runtime Runtime) error {
	envDir := filepath.Join(baseDir, "environments", env)
	resourceFile := filepath.Join(envDir, "resources.yaml")
	resourceContents := fmt.Sprintf(`context: %s
region: %s
state_store: %s
`, context, region, fmt.Sprintf("xlrte-state-%s-%s", context, secrets.RandStringBytesOfLength(6)))
	err := os.MkdirAll(filepath.Clean(envDir), 0750)
	if err != nil {
		return err
	}
	_, err = os.Stat(resourceFile)
	if err != nil {
		_, err = os.Create(filepath.Clean(resourceFile))
		if err != nil {
			return err
		}
		err = os.WriteFile(resourceFile, []byte(resourceContents), 0600)
		if err != nil {
			return err
		}
	}

	return runtime.InitEnvironment(ctx, env, context, region)
}

func Prepare(rootDir string, selector EnvResolver, runtimes *Runtimes) ([]*DeploymentConfig, preApplyFn, error) {
	configs, err := parseDeploymentConfig(rootDir, selector, runtimes)
	if err != nil {
		return nil, nil, err
	}
	preFn, err := configureDeployment(rootDir, configs)
	if err != nil {
		return nil, nil, err
	}
	return configs, preFn, nil
}

func Execute(ctx context.Context, cmd Command, rootDir string, selector EnvResolver, runtimes *Runtimes) error {
	configs, _, err := Prepare(rootDir, selector, runtimes)
	if err != nil {
		return err
	}
	for _, config := range configs {
		err := exec(ctx, cmd, config.Runtime)
		if err != nil {
			return err
		}
	}
	return nil
}

func exec(ctx context.Context, cmd Command, rte Runtime) error {
	switch cmd {
	case Plan:
		return rte.Plan(ctx)
	case Export:
		return rte.Export(ctx)
	case Apply:
		return rte.Apply(ctx)
	case Delete:
		return rte.Delete(ctx)
	}
	return fmt.Errorf("no matching command found")
}

func configureDeployment(baseDir string, deployments []*DeploymentConfig) (preApplyFn, error) {
	outputs := &EnvVars{}
	toApply := []preApplyFn{}
	var err error
	dependencyDefinitions := []DependencyBinding{}
	resources := []Resource{}

	for _, deployment := range deployments {
		tmpResources := []Resource{}

		for _, loader := range deployment.Runtime.Resources() {
			for _, defs := range deployment.Resources {

				if defs.Name == loader.Name() {
					rs, bindings, e := loader.Load(defs)
					if e != nil {
						return nil, e
					}
					dependencyDefinitions = append(dependencyDefinitions, bindings...)
					tmpResources = append(tmpResources, rs...)
				}
			}
		}

		added := make(map[ResourceIdentity]string)
		for _, r := range tmpResources {
			resourceID := r.Identity()
			if added[resourceID] == "" {
				added[resourceID] = resourceID.ID
				resources = append(resources, r)
			}
		}
		deployment.underlyingResources = resources
		for _, resource := range resources {
			migrate, ok := resource.(CanMigrate)
			if ok {
				toApply = append(toApply, migrate.Migrate)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	for _, deployment := range deployments {
		envCtx := deployment.Environment.ctx()
		err = deployment.Runtime.Init(envCtx)
		if err != nil {
			return nil, err
		}
		for _, service := range deployment.Services {
			envKeys := []string{}
			refKeys := []string{}
			secretKeys := []string{}
			for k := range service.Env.Vars {
				envKeys = append(envKeys, k)
			}
			for k := range service.Env.Refs {
				refKeys = append(refKeys, k)
			}
			for k := range service.Env.Secrets {
				secretKeys = append(secretKeys, k)
			}
			outputs.Merge(service.Env)
			outputs.Merge(deployment.Environment.Env)
			serviceResources := deployment.Environment.Resources[service.Runtime]
			var resourceBytes *[]byte
			if serviceResources != nil {
				marshalled, e := yaml.Marshal(deployment.Environment.Resources[service.Runtime])
				if e != nil {
					return nil, e
				}
				resourceBytes = &marshalled
			}
			envVars := outputs.withKeys(envKeys, refKeys, secretKeys)
			deploymentContext := DeploymentContext{Env: envVars, Resources: resourceBytes}
			for _, serviceLoader := range deployment.Runtime.Services() {
				resource, e := serviceLoader.Load(envCtx, service, deploymentContext)
				if e != nil {
					return nil, e
				}
				deployment.underlyingResources = append(deployment.underlyingResources, resource)
				resources = append(resources, resource)
			}
		}
		homeDir, e := os.UserHomeDir()
		if e != nil {
			return nil, e
		}
		allSecrets, e := secrets.GetAllSecrets(homeDir, baseDir, envCtx.EnvName)
		if e != nil {
			return nil, e
		}
		for _, dep := range dependencyDefinitions {
			if dep.SecretRefs != nil && len(dep.SecretRefs) > 0 {
				for _, ref := range dep.SecretRefs {
					ref.Name = fmt.Sprintf("%s_%s", dep.Identity.String(), ref.Name)
					foundSecret := false
					for _, secret := range allSecrets {
						if secret.Name == ref.Name {
							foundSecret = true
							break
						}
					}
					if !foundSecret {
						newSecret := ref.Generate()
						allSecrets = append(allSecrets, newSecret)
						err = secrets.WriteSecret(baseDir, envCtx.EnvName, newSecret.Name, newSecret.Value)
						if err != nil {
							return nil, err
						}
					}
				}
			}
		}

		err = deployment.Runtime.InitSecrets(envCtx, allSecrets)
		if err != nil {
			return nil, err
		}
	}

	for _, resource := range resources {
		for _, dep := range dependencyDefinitions {
			if resource.Identity() == dep.DependedOnBy && dep.Config != nil {
				e := dep.Config.ConfigureResource(resource)
				if e != nil {
					return nil, e
				}
			}
		}
		err = resource.Configure()
		if err != nil {
			return nil, err
		}
	}

	return func(ctx context.Context) error {
		for _, f := range toApply {
			e := f(ctx)
			if e != nil {
				return e
			}
		}
		return nil
	}, nil
}

func parseDeploymentConfig(rootDir string, selector EnvResolver, runtimes *Runtimes) ([]*DeploymentConfig, error) {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("The directory " + rootDir + " does not exist")
	}
	envs, err := ReadAllEnvironments(filepath.Join(rootDir, "environments"))
	if err != nil {
		return nil, err
	}
	svcs, err := ReadAllServices(filepath.Join(rootDir, "services"))
	if err != nil {
		return nil, err
	}
	targetEnv := Environment{}
	for i := range envs {
		if selector.Env() == envs[i].EnvName {
			targetEnv = envs[i]
			targetEnv.Resolver = selector
		}
	}
	if targetEnv.Name() == "" {
		return nil, fmt.Errorf("could not find a target environment for %v", selector.Env())
	}
	if len(svcs) == 0 {
		return nil, fmt.Errorf("no services to deploy")
	}

	runtimeMap := make(map[string]*DeploymentConfig)
	for _, svc := range svcs {
		rtme, err := runtimes.getRuntimeFor(svc)
		if err != nil {
			return nil, err
		}
		conf := runtimeMap[rtme.Name()]
		if conf == nil {
			conf = &DeploymentConfig{
				Runtime:     rtme,
				Services:    []*Service{svc},
				Environment: &targetEnv,
			}
			runtimeMap[rtme.Name()] = conf
		} else {
			conf.Services = append(conf.Services, svc)
		}

		usedKeys := []string{}
		for k, v := range svc.DependsOn {
			_, err := supportsResourceType(rtme, svc.Name(), k)
			if err != nil {
				return nil, err
			}
			bytes, err := yaml.Marshal(v)
			if err != nil {
				return nil, err
			}
			usedKeys = append(usedKeys, k)
			inf := targetEnv.Resources[k]
			if inf != nil {
				resourceBytes, err := yaml.Marshal(inf)
				if err != nil {
					return nil, err
				}
				resource := ResourceDefinition{DependedOnBy: svc.ToIdentity(), ServiceConfig: bytes, ResourceConfig: &resourceBytes, Name: k}
				conf.Resources = append(conf.Resources, &resource)
			} else {
				resource := ResourceDefinition{DependedOnBy: svc.ToIdentity(), ServiceConfig: bytes, ResourceConfig: nil, Name: k}
				conf.Resources = append(conf.Resources, &resource)
			}
		}
		unclaimedResources := make(map[string]interface{})
		for k, v := range targetEnv.Resources {
			found := false
			for _, usedKey := range usedKeys {
				if k == usedKey {
					found = true
					break
				}
			}
			if !found {
				unclaimedResources[k] = v
			}
		}
		for _, r := range conf.Resources {
			r.unclaimedResources = unclaimedResources
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
		for _, loader := range rtime.Services() {
			if loader.Name() == service.Runtime {
				return rtime, nil
			}
		}
	}
	return nil, fmt.Errorf("could not find a Runtime that supports service of type %s", service.Runtime)
}
