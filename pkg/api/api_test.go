package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Read_Service_Definition(t *testing.T) {
	service, err := ReadServiceDefinition("testdata/cloudrun-srv.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "cloudrun-srv", service.Name)
	assert.Equal(t, "cloudrun", service.Runtime)
	assert.Equal(t, "gcr.io/chaordic/hello-app", service.Artifact.BaseName)
	assert.Equal(t, "docker", service.Artifact.Type)
}

func Test_Invalid_Service(t *testing.T) {
	_, err := ReadServiceDefinition("testdata/invalid-service.yaml")

	assert.Error(t, err)
	validations := GetValidationErrors(err)
	assert.NotNil(t, validations)
	assert.True(t, len(*validations) > 0)

	_, err = ReadServiceDefinition("testdata/invalid-service.bla")

	assert.Error(t, err)
	validations = GetValidationErrors(err)
	assert.Nil(t, validations)
}

func Test_Load_Plugins(t *testing.T) {
	runtimes, err := GetRuntimes("testdata/plugins")

	assert.NoError(t, err)
	assert.Len(t, runtimes, 1)
}
