package gcp

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api"
)

func Test_CloudStorage_Can_Read_Config(t *testing.T) {
	serviceData := getCloudRunBytes(t, filepath.Join("testdata", "cloudstorage", "service.yaml"), "cloudstorage")
	confData := getCloudRunBytes(t, filepath.Join("testdata", "cloudstorage", "resources.yaml"), "cloudstorage")

	resource := &gcsConfig{}
	resources, bindings, err := resource.Load(&api.ResourceDefinition{
		Name:           "pubsub",
		DependedOnBy:   api.ResourceIdentity{ID: "the-service", Type: "cloudrun"},
		ServiceConfig:  serviceData,
		ResourceConfig: &confData,
	})
	assert.NoError(t, err)
	assert.Len(t, bindings, 4)
	assert.Len(t, resources, 2)

	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "bar"}, resources[0].Identity())
	assert.Equal(t, "EU", resources[0].(*gcsConfig).Location)
	assert.Equal(t, "MULTI_REGIONAL", resources[0].(*gcsConfig).StorageClass)
	assert.Equal(t, true, *resources[0].(*gcsConfig).VersioningEnabled)

	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "baz"}, resources[1].Identity())
	assert.Equal(t, "US", resources[1].(*gcsConfig).Location)
	assert.Equal(t, "STANDARD", resources[1].(*gcsConfig).StorageClass)
	assert.Equal(t, false, *resources[1].(*gcsConfig).VersioningEnabled)

	assert.Equal(t, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"}, bindings[0].DependedOnBy)
	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "foo-bucket"}, bindings[0].Identity)
	assert.Equal(t, api.ReadOnly, bindings[0].Privileges)
	assert.Equal(t, &gcsIAM{"foo-bucket", "roles/storage.objectViewer"}, bindings[0].Config)

	assert.Equal(t, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"}, bindings[1].DependedOnBy)
	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "bar"}, bindings[1].Identity)
	assert.Equal(t, api.Owner, bindings[1].Privileges)
	assert.Equal(t, &gcsIAM{"bar", "roles/storage.objectAdmin"}, bindings[1].Config)

	assert.Equal(t, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"}, bindings[2].DependedOnBy)
	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "baz"}, bindings[2].Identity)
	assert.Equal(t, api.Owner, bindings[2].Privileges)
	assert.Equal(t, &gcsIAM{"baz", "roles/storage.objectAdmin"}, bindings[2].Config)

	assert.Equal(t, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"}, bindings[3].DependedOnBy)
	assert.Equal(t, api.ResourceIdentity{Type: "cloudstorage", ID: "bazf"}, bindings[3].Identity)
	assert.Equal(t, api.ReadWrite, bindings[3].Privileges)
	assert.Equal(t, &gcsIAM{"bazf", "roles/storage.objectAdmin"}, bindings[3].Config)

}

func Test_IAM_Configures_Resource(t *testing.T) {
	iam := &gcsIAM{"baz", "roles/storage.objectAdmin"}
	cloudRun := cloudRunConfig{}

	err := iam.ConfigureResource(&cloudRun)
	assert.NoError(t, err)
	assert.Equal(t, cloudRun.DependsOn, []string{"module.cloudstorage-baz.bucket"})
	assert.Equal(t, cloudRun.CloudStorage, []*gcsIAM{iam})

}
