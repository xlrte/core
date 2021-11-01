package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type validations struct {
	fileLocation string
	err          error
}

type ConfigProvider = func(dir string, file string) Named

func (e *validations) Error() string {
	return fmt.Sprintf("validation in config %s, errors: %v", e.fileLocation, e.err)
}

func readDefinition(fileLocation string, strct interface{}) error {
	data, err := ioutil.ReadFile(filepath.Clean(fileLocation))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(data), strct)
	if err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(strct); err != nil {
		// TODO: improve error messages here
		return &validations{
			fileLocation: fileLocation,
			err:          err,
		}
	}
	return nil
}

func readAllDefinitions(serviceDir string, createFn ConfigProvider) ([]Named, error) {
	defs := []Named{}
	if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("The directory " + serviceDir + " does not exist")
	}
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || (!strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml")) {
			return nil
		}
		dir, _ := filepath.Split(path)
		parent := filepath.Base(dir)
		def := createFn(parent, info.Name())
		if def == nil {
			return nil
		}
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
	switch e := err.(type) { // nolint
	case validator.ValidationErrors:
		return &e
	case *validations:
		return GetValidationErrors(e.err)
	default:
		return nil
	}
}

func ReadAllServices(serviceDir string) ([]*Service, error) {
	defs, err := readAllDefinitions(serviceDir, func(dir, file string) Named {
		return &Service{}
	})
	if err != nil {
		return nil, err
	}
	services := []*Service{}
	for _, def := range defs {
		services = append(services, def.(*Service))
	}
	return services, nil
}

func ReadAllEnvironments(envDir string) ([]Environment, error) {
	defs, err := readAllDefinitions(envDir, func(dir, file string) Named {
		if (file != "resources.yml" && file != "resources.yaml") || dir == "" {
			return nil
		}
		return &Environment{
			EnvName: dir,
		}
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
