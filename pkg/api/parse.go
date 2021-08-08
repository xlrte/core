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
