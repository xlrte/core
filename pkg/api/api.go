package api

type Runtimes struct {
	Runtimes []Runtime
}

type DeploymentConfig struct {
	Environment *Environment
	Runtime     Runtime
	Services    []*Service
	Resources   []*ResourceDefinition
}

// Resources have an internal dependency order with outputs.
type Runtime interface {
	Named
	SupportServiceRuntime(name string) bool
	Resources() []ResourceLoader
	Configure(name string, artifact Artifact, env EnvVars) error
}

type ResourceDefinition struct {
	name           string
	serviceConfig  []byte
	resourceConfig *[]byte
}

type ResourceLoader interface {
	Named
	Load(definition ResourceDefinition) ([]*Resource, error)
}

type Resource interface {
	IsSame(resource Resource) (bool, error)
	Apply(previousOutputs Outputs) (Outputs, error)
	Migrate() error
}

type Outputs struct {
	Values map[string]interface{}
}

type Named interface {
	Name() string
}

type TriggerSelector interface {
	IsMatch(env Environment) bool
}

type NameSelector struct {
	EnvName string
}

func (name *NameSelector) IsMatch(env Environment) bool {
	return env.Name() == name.EnvName
}

func (env *Environment) Name() string {
	return env.EnvName
}

func (svc *Service) Name() string {
	return svc.SVCName
}
