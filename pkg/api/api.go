package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

type Runtimes struct {
	Runtimes []Runtime
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

func readServiceDefinition(fileLocation string) (Service, error) {
	var service Service
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return Service{}, err
	}
	err = yaml.Unmarshal([]byte(data), &service)
	if err != nil {
		return Service{}, err
	}

	validate := validator.New()
	if errs := validate.Struct(service); errs != nil {
		return Service{}, errs
	}
	return service, nil
}

func GetValidationErrors(err error) *validator.ValidationErrors {
	switch e := err.(type) {
	case validator.ValidationErrors:
		return &e
	default:
		return nil
	}
}

func (runtimes *Runtimes) GetRuntimeFor(service *Service) (Runtime, error) {
	for _, rtime := range runtimes.Runtimes {
		if rtime.SupportServiceRuntime(service.Runtime) {
			return rtime, nil
		}
	}
	return nil, fmt.Errorf("could not find a Runtime that supports service of type %s", service.Runtime)
}

func ReadAllServices(serviceDir string) ([]Service, error) {
	services := []Service{}

	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return nil
		}
		service, e := readServiceDefinition(path)
		if e != nil {
			return e
		}
		services = append(services, service)

		return nil
	})
	if err != nil {
		return nil, err
	}
	names := make(map[string]*Service)
	for _, service := range services {
		_, found := names[service.Name]
		if found {
			return nil, fmt.Errorf("duplicate service with name %s found", service.Name)
		}
		names[service.Name] = &service
	}

	return services, nil
}
