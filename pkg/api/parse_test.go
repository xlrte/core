package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Read_Service_Definition(t *testing.T) {
	service := &Service{}
	err := readDefinition("testdata/cloudrun-srv.yaml", service)

	assert.NoError(t, err)
	assert.Equal(t, "cloudrun-srv", service.Name())
	assert.Equal(t, "cloudrun", service.Runtime)
	assert.NotNil(t, service.Spec)
}

func Test_Invalid_Service(t *testing.T) {
	service := &Service{}
	err := readDefinition("testdata/invalid-service.yaml", service)

	assert.Error(t, err)
	validations := GetValidationErrors(err)
	assert.NotNil(t, validations)
	assert.True(t, len(*validations) > 0)

	err = readDefinition("testdata/invalid-service.bla", service)

	assert.Error(t, err)
	validations = GetValidationErrors(err)
	assert.Nil(t, validations)
}

func Test_ReadAll_Service_Definitions(t *testing.T) {
	services, err := ReadAllServices("testdata/services")

	assert.NoError(t, err)
	assert.Len(t, services, 2)

	names := make(map[string]*string)

	for _, svc := range services {
		names[svc.SVCName] = &svc.SVCName // nolint
	}
	assert.NotNil(t, names["cloudrun-srv"])
	assert.NotNil(t, names["cloudrun-srv2"])
}

func Test_ReadAll_Service_Definitions_Duplicate_Name(t *testing.T) {
	_, err := ReadAllServices(filepath.Join("testdata", "duplicate-services"))
	assert.Error(t, err)
}

func Test_Read_Env_Definition(t *testing.T) {
	envs, err := ReadAllEnvironments(filepath.Join("testdata", "environments"))
	assert.NoError(t, err)

	assert.Equal(t, len(envs), 1)
	assert.Equal(t, envs[0].EnvName, "prod")
	assert.Equal(t, envs[0].Region, "europe-west6")
}

func Test_Merge_EnvVars(t *testing.T) {
	vars := EnvVars{Vars: map[string]string{
		"foo": "bar",
		"bar": "baz",
	}, Refs: map[string]string{
		"ref":  "baz",
		"ref2": "baz",
	},
		Secrets: map[string]string{
			"sec":  "baz",
			"sec2": "baz",
		},
	}

	toAdd := EnvVars{Vars: map[string]string{
		"foo": "baz",
	}, Refs: map[string]string{
		"ref": "q",
	},
		Secrets: map[string]string{
			"sec": "q",
		},
	}

	expected := EnvVars{Vars: map[string]string{
		"foo": "baz",
		"bar": "baz",
	}, Refs: map[string]string{
		"ref":  "q",
		"ref2": "baz",
	},
		Secrets: map[string]string{
			"sec":  "q",
			"sec2": "baz",
		},
	}

	vars.Merge(toAdd)

	assert.Equal(t, vars, expected)

}
