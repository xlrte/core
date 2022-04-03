package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xlrte/core/pkg/api/secrets"
	"gopkg.in/yaml.v2"
)

type Runtimes struct {
	Runtimes []Runtime
}

type DependencyPrivileges = int
type ConfigType = int
type SecretType = int

const (
	ReadOnly DependencyPrivileges = iota
	ReadWrite
	Owner
)

const (
	String ConfigType = iota
	Bool
	Number
)

const (
	RandomString SecretType = iota
)

type SecretRef struct {
	Name string
	Type SecretType
}

type DeploymentConfig struct {
	Environment         *Environment
	Runtime             Runtime
	Services            []*Service
	Resources           []*ResourceDefinition
	underlyingResources []Resource
}

// Resources have an internal dependency order with outputs.
type Runtime interface {
	Named
	//InitEnvironment initialises an environment for the first time, such as 'dev', 'prod' etc.
	InitEnvironment(ctx context.Context, env, project, region string) error
	//Init initialises for a plan or apply
	Init(env EnvContext) error
	//InitSecrets initialises the secrets system.
	InitSecrets(env EnvContext, secrets []*secrets.Secret) error
	Resources() []ResourceLoader
	Services() []ServiceLoader
	Apply(ctx context.Context) error
	Plan(ctx context.Context) error
	Delete(ctx context.Context) error
	Export(ctx context.Context) error
}

type ResourceDefinition struct {
	Name               string
	DependedOnBy       ResourceIdentity
	ServiceConfig      []byte
	ResourceConfig     *[]byte
	unclaimedResources map[string]interface{}
}

type ResourceLoader interface {
	Named
	Load(definition *ResourceDefinition) ([]Resource, []DependencyBinding, error)
}

type ServiceLoader interface {
	Named
	Load(envCtx EnvContext, service *Service, deploymentContext DeploymentContext) (Resource, error)
}

type Resource interface {
	Identity() ResourceIdentity
	Configure() error
}

type CanMigrate interface {
	Migrate(ctx context.Context) error
}

// ResourceIdentity is something that uniquely identifies instances of a Resource, for instance Type: "cloudsql", ID: "the-database-name"
type ResourceIdentity struct {
	Type string
	ID   string
}
type DependencyBinding struct {
	DependedOnBy ResourceIdentity
	Privileges   DependencyPrivileges
	Identity     ResourceIdentity
	Config       DependencyVisitor
	SecretRefs   []SecretRef
}

type DependencyVisitor interface {
	ConfigureResource(resource Resource) error
}

type DeploymentContext struct {
	Env       EnvVars
	Resources *[]byte
}

type Named interface {
	Name() string
}

type EnvResolver interface {
	Version(serviceName string) (string, error)
	Env() string
}

type ArgEnvResolver struct {
	EnvName, ArgVersion string
}

type FileEnvResolver struct {
	EnvName  string
	BaseDir  string
	versions map[string]string
}

func (name *ArgEnvResolver) Version(serviceName string) (string, error) {
	return name.ArgVersion, nil
}

func (name *ArgEnvResolver) Env() string {
	return name.EnvName
}

func (name *FileEnvResolver) Env() string {
	return name.EnvName
}

func (name *FileEnvResolver) Version(serviceName string) (string, error) {
	if name.versions == nil {
		versions := make(map[string]string)
		dir := filepath.Join(name.BaseDir, "environments", name.EnvName)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			thePath := strings.ToLower(path)
			if (strings.HasSuffix(thePath, "versions.yaml") || strings.HasSuffix(thePath, "versions.yml")) && name.versions == nil {
				data, err := ioutil.ReadFile(filepath.Clean(path))
				if err != nil {
					return err
				}
				err = yaml.Unmarshal(data, &versions)
				if err != nil {
					return err
				}
				name.versions = versions
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		name.versions = versions
	}
	version := name.versions[serviceName]
	if version != "" {
		return version, nil
	}
	return "", fmt.Errorf("could not resolve a version for service `%s` in env `%s`. Have you added the service to the `%s/environments/%s/version.yaml` file?\n It should have the format `%s: [version]`", serviceName, name.Env(), name.BaseDir, name.Env(), serviceName)
}

func (env *Environment) Name() string {
	return env.EnvName
}

func (svc *Service) Name() string {
	return svc.SVCName
}

func (id ResourceIdentity) String() string {
	return fmt.Sprintf("%s-%s", id.Type, id.ID)
}

func (rd *ResourceDefinition) GetConfig(name string, readInto interface{}) error {
	if rd.unclaimedResources[name] == nil {
		return nil
	}
	bytes, err := yaml.Marshal(rd.unclaimedResources[name])
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, readInto)
	return err
}

func (secretRef *SecretRef) Generate() *secrets.Secret {
	sec := &secrets.Secret{Name: secretRef.Name, Value: secrets.RandStringBytes()}
	return sec
}
