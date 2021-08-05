package api

type Service struct {
	Name      string                 `yaml:"name" validate:"required"`
	Runtime   string                 `yaml:"runtime" validate:"oneof=cloudrun lambda k8s"`
	Artifact  Artifact               `yaml:"artifact" validate:"required"`
	Network   map[string]interface{} `yaml:"network"`
	DependsOn map[string]interface{} `yaml:"depends_on"`
	Env       Env                    `yaml:"env"`
}

type Artifact struct {
	OriginRepo string `yaml:"originRepo"`
	BaseName   string `yaml:"base_name" validate:"required"`
	Type       string `yaml:"type" validate:"oneof=docker zip"`
}

type Http struct {
	Public bool   `yaml:"public" validate:"required"`
	Path   string `yaml:"path"`
}

type Env struct {
	Vars map[string]string `yaml:"vars"`
	// Secrets
}

type Output struct {
	Key   string
	Value string
}

// Resources have an internal dependency order with outputs.
type Runtime interface {
	Runtime() string
	Resources() []*Resource
	Configure(name string, artifact Artifact, env Env, previousOutputs []Output) error
}

type Resource interface {
	IsMatch(name string) bool
	Configure(serviceConfig []byte, resourceConfig []byte, previousOutputs []Output) ([]Output, error)
	Migrate() error
}
