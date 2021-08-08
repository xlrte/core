package api

// Resources have an internal dependency order with outputs.
type Runtime interface {
	SupportServiceRuntime(name string) bool
	Resources() []Resource
	Configure(name string, artifact Artifact, env EnvVars) error
}

type Resource interface {
	Sort(resources []Resource) []Resource
	Name() string
	Configure(serviceConfig []byte, resourceConfig []byte) error
	Migrate() error
}
