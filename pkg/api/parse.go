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

func readDefinition(fileLocation string, strct interface{}) error {
	data, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(data), strct)
	if err != nil {
		return err
	}

	validate := validator.New()
	if errs := validate.Struct(strct); errs != nil {
		return errs
	}
	return nil
}

func readAllDefinitions(serviceDir string, createFn func() Named) ([]Named, error) {
	defs := []Named{}

	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return nil
		}
		def := createFn()
		e := readDefinition(path, def)
		if e != nil {
			return e
		}
		defs = append(defs, def)

		return nil
	})
	if err != nil {
		return nil, err
	}
	names := make(map[string]Named)
	for _, def := range defs {
		_, found := names[def.Name()]
		if found {
			return nil, fmt.Errorf("duplicate definition with name %s found", def.Name())
		}
		names[def.Name()] = def
	}

	return defs, nil
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
	defs, err := readAllDefinitions(serviceDir, func() Named {
		return &Service{}
	})
	if err != nil {
		return nil, err
	}
	services := []Service{}
	for _, def := range defs {
		services = append(services, *def.(*Service))
	}
	return services, nil
}

func ReadAllEnvironments(envDir string) ([]Environment, error) {
	defs, err := readAllDefinitions(envDir, func() Named {
		return &Environment{}
	})
	if err != nil {
		return nil, err
	}
	services := []Environment{}
	for _, def := range defs {
		services = append(services, *def.(*Environment))
	}
	return services, nil
}
