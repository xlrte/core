package api

type Environment struct {
	EnvName    string                 `yaml:"name" validate:"required"`
	IsDynamic  bool                   `yaml:"dynamic"`
	Provider   string                 `yaml:"provider" validate:"required"`
	Context    string                 `yaml:"context" validate:"required"`
	Region     string                 `yaml:"region" validate:"required"`
	Resources  map[string]interface{} `yaml:"resources"`
	Deployment Deployment             `yaml:"deployment" validate:"required"`
	Env        EnvVars                `yaml:"env"`
}

type Deployment struct {
	Trigger DeploymentTrigger `yaml:"trigger" validate:"required"`
}

type ServiceRepo struct {
	Repo        string `yaml:"repo" validate:"required"`
	ServiceName string `yaml:"service" validate:"required"`
}

type DeploymentTrigger struct {
	IsLocal bool       `yaml:"local"`
	Git     GitTrigger `yaml:"git"`
}

type LocalTrigger struct{}

type GitTrigger struct {
	BranchPattern string        `yaml:"branch" validate:"required"`
	Event         string        `yaml:"event" validate:"oneof=commit tag pr"`
	Repos         []ServiceRepo `yaml:"repos"`
}

type Service struct {
	SVCName   string                 `yaml:"name" validate:"required"`
	Runtime   string                 `yaml:"runtime" validate:"required"`
	Artifact  Artifact               `yaml:"artifact" validate:"required"`
	Network   map[string]interface{} `yaml:"network"`
	DependsOn map[string]interface{} `yaml:"depends_on"`
	Env       EnvVars                `yaml:"env"`
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

type EnvVars struct {
	Vars map[string]string `yaml:"vars"`
	// Secrets
}
