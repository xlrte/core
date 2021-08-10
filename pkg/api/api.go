package api

// Resources have an internal dependency order with outputs.
type Runtime interface {
	SupportServiceRuntime(name string) bool
	Resources() []Resource
	Configure(name string, artifact Artifact, env EnvVars) error
}

type Resource interface {
	Named
	Sort(resources []Resource) []Resource
	Configure(serviceConfig []byte, resourceConfig []byte) error
	Migrate() error
}

type Named interface {
	Name() string
}

func (env *Environment) Name() string {
	return env.EnvName
}

func (svc *Service) Name() string {
	return svc.SVCName
}
