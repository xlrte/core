package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Read_Service_Definition(t *testing.T) {
	service, err := readServiceDefinition("testdata/cloudrun-srv.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "cloudrun-srv", service.Name)
	assert.Equal(t, "cloudrun", service.Runtime)
	assert.Equal(t, "gcr.io/chaordic/hello-app", service.Artifact.BaseName)
	assert.Equal(t, "docker", service.Artifact.Type)
}

func Test_Invalid_Service(t *testing.T) {
	_, err := readServiceDefinition("testdata/invalid-service.yaml")

	assert.Error(t, err)
	validations := GetValidationErrors(err)
	assert.NotNil(t, validations)
	assert.True(t, len(*validations) > 0)

	_, err = readServiceDefinition("testdata/invalid-service.bla")

	assert.Error(t, err)
	validations = GetValidationErrors(err)
	assert.Nil(t, validations)
}

func Test_ReadAll_Service_Definitions(t *testing.T) {
	services, err := ReadAllServices("testdata/services")

	assert.NoError(t, err)
	assert.Len(t, services, 2)
}

func Test_ReadAll_Service_Definitions_Duplicate_Name(t *testing.T) {
	_, err := ReadAllServices("testdata/duplicate-services")

	assert.Error(t, err)
}
