package api

type Environment struct {
	Context    string `yaml:"context" validate:"required"`
	RepoBase   string `docker_repo:"context"`
	Region     string `yaml:"region" validate:"required"`
	StateStore string `yaml:"state_store" validate:"required"`
	EnvName    string `yaml:"-" validate:"required"`
	//	Provider   string                 `yaml:"provider" validate:"required"`
	Resources map[string]interface{} `yaml:"resources"`
	// Deployment Deployment             `yaml:"deployment" validate:"required"`
	Env      EnvVars     `yaml:"env"`
	Resolver EnvResolver `yaml:"-"`
}

type EnvContext struct {
	Context    string
	RepoBase   string
	Region     string
	StateStore string
	EnvName    string
	Version    func(string) (string, error)
}

type Service struct {
	SVCName   string                 `yaml:"name" validate:"required"`
	Runtime   string                 `yaml:"runtime" validate:"required"`
	Spec      interface{}            `yaml:"spec" validate:"required"`
	DependsOn map[string]interface{} `yaml:"depends_on"`
	Env       EnvVars                `yaml:"env"`
}

type EnvVars struct {
	Vars    map[string]string `yaml:"vars"`
	Secrets map[string]string `yaml:"secrets"`
	Refs    map[string]string `yaml:"-"`
}

func (service *Service) ToIdentity() ResourceIdentity {
	return ResourceIdentity{Type: service.Runtime, ID: service.SVCName}
}

// This only needs to merge Vars, as other are Refs from the env.
func (vars *EnvVars) withKeys(varKeys []string, refKeys []string, secretKeys []string) EnvVars {
	theVars := make(map[string]string)
	theRefs := make(map[string]string)
	theSecrets := make(map[string]string)
	for k, v := range vars.Vars {
		for _, key := range varKeys {
			if key == k {
				theVars[k] = v
				break
			}
		}
	}
	for k, v := range vars.Refs {
		for _, key := range refKeys {
			if key == k {
				theRefs[k] = v
				break
			}
		}
	}
	for k, v := range vars.Secrets {
		for _, key := range secretKeys {
			if key == k {
				theSecrets[k] = v
				break
			}
		}
	}
	return EnvVars{Vars: theVars, Refs: theRefs, Secrets: theSecrets}
}

func (env *Environment) ctx() EnvContext {
	return EnvContext{
		Context:    env.Context,
		Region:     env.Region,
		EnvName:    env.EnvName,
		StateStore: env.StateStore,
		Version:    env.Resolver.Version,
		RepoBase:   env.RepoBase,
	}
}

func (output *EnvVars) Merge(secondOutput EnvVars) {
	if output.Vars == nil {
		output.Vars = secondOutput.Vars
	} else if secondOutput.Vars != nil {
		for k, v := range secondOutput.Vars {
			output.Vars[k] = v
		}
	}

	if output.Refs == nil {
		output.Refs = secondOutput.Refs
	} else if secondOutput.Refs != nil {
		for k, v := range secondOutput.Refs {
			output.Refs[k] = v
		}
	}

	if output.Secrets == nil {
		output.Secrets = secondOutput.Secrets
	} else if secondOutput.Secrets != nil {
		for k, v := range secondOutput.Secrets {
			output.Secrets[k] = v
		}
	}
}
