package api

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v2"
)

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
	SupportServiceRuntime(name string) bool
	Resources() []Resource
	Configure(name string, artifact Artifact, env Env, previousOutputs []Output) error
}

type Resource interface {
	Sort(resources []Resource) []Resource
	Name() string
	Configure(serviceConfig []byte, resourceConfig []byte, previousOutputs []Output) ([]Output, error)
	Migrate() error
}

func ReadServiceDefinition(fileLocation string) (*Service, error) {
	var service Service
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(data), &service)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	if errs := validate.Struct(service); errs != nil {
		return nil, errs
	}
	return &service, nil
}

func GetValidationErrors(err error) *validator.ValidationErrors {
	switch e := err.(type) {
	case validator.ValidationErrors:
		return &e
	default:
		return nil
	}
}

func GetRuntimeFor(service *Service) *Runtime {
	return nil
}

func GetRuntimes(basePath string) ([]Runtime, error) {
	runtimes := []Runtime{}
	err := filepath.WalkDir(basePath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.Type().IsRegular() || !strings.HasSuffix(path, ".so") {
			return nil
		}

		plug, err := plugin.Open(path)
		if err != nil {
			fmt.Println("plugin.Open")
			fmt.Println(err)
			return err
		}

		symRuntime, err := plug.Lookup("Runtime")
		if err != nil {

			fmt.Println("plugin.Loopup")
			fmt.Println(err)
			return err
		}
		var runtime Runtime
		runtime, ok := symRuntime.(Runtime)
		if !ok {

			fmt.Println("plugin.Cast")
			fmt.Println(ok)
			return fmt.Errorf("the plugin %s is not a valid implementation of Runtime", path)
		}
		runtimes = append(runtimes, runtime)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return runtimes, nil
}
