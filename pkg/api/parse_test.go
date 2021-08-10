package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Read_Service_Definition(t *testing.T) {
	service := &Service{}
	err := readDefinition("testdata/cloudrun-srv.yaml", service)

	assert.NoError(t, err)
	assert.Equal(t, "cloudrun-srv", service.Name())
	assert.Equal(t, "cloudrun", service.Runtime)
	assert.Equal(t, "gcr.io/chaordic/hello-app", service.Artifact.BaseName)
	assert.Equal(t, "docker", service.Artifact.Type)
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
		names[svc.SVCName] = &svc.SVCName
	}
	assert.NotNil(t, names["cloudrun-srv"])
	assert.NotNil(t, names["cloudrun-srv2"])
}

func Test_ReadAll_Service_Definitions_Duplicate_Name(t *testing.T) {
	_, err := ReadAllServices("testdata/duplicate-services")
	assert.Error(t, err)
}
