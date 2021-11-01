package gcp

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

func Test_Can_Read_Resources(t *testing.T) {
	serviceData := getCloudRunBytes(t, filepath.Join("testdata", "cloudsql", "service.yaml"), "cloudsql")
	confData := getCloudRunBytes(t, filepath.Join("testdata", "cloudsql", "resources.yaml"), "cloudsql")

	resource := &cloudSql{}
	resources, bindings, err := resource.Load(&api.ResourceDefinition{
		Name:           "cloudsql",
		DependedOnBy:   api.ResourceIdentity{ID: "the-service", Type: "cloudrun"},
		ServiceConfig:  serviceData,
		ResourceConfig: &confData,
	})

	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Len(t, bindings, 3)

	assert.Equal(t, bindings[1].DependedOnBy, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"})
	assert.Equal(t, bindings[1].Identity, api.ResourceIdentity{Type: "cloudsql", ID: "my-pg-db"})
	assert.Equal(t, bindings[1].Privileges, api.Owner)
	// assert.Equal(t, bindings[0].Outputs, []string{"connection_name"})

	db := resources[1].(*cloudSql)

	assert.Equal(t, db.DbName, "my-pg-db")
	assert.Equal(t, db.DBType, "POSTGRES_13")
	assert.Equal(t, db.Size, 10)
	assert.Equal(t, db.MachineType, "db-f1-micro")
	assert.Equal(t, db.DeleteProtection, true)
}

func getCloudRunBytes(t *testing.T, path string, mapPath string) []byte {
	serviceData, err := ioutil.ReadFile(filepath.Clean(path))
	assert.NoError(t, err)

	var theMap map[string]interface{}
	err = yaml.Unmarshal(serviceData, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap[mapPath])
	sBytes, err := yaml.Marshal(theMap[mapPath])
	assert.NoError(t, err)
	return sBytes
}

func Test_CloudSqlTemplate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	serviceData := getCloudRunBytes(t, filepath.Join("testdata", "cloudsql", "service.yaml"), "cloudsql")
	confData := getCloudRunBytes(t, filepath.Join("testdata", "cloudsql", "resources.yaml"), "cloudsql")

	resource := &cloudSql{baseDir: tmpDir}
	resources, deps, err := resource.Load(&api.ResourceDefinition{
		Name:           "cloudsql",
		DependedOnBy:   api.ResourceIdentity{ID: "the-service", Type: "cloudrun"},
		ServiceConfig:  serviceData,
		ResourceConfig: &confData,
	})
	assert.NoError(t, err)
	err = resources[1].Configure()

	assert.Equal(t, len(deps), 3)
	assert.Equal(t, deps[2].Privileges, api.Owner)
	// assert.Equal(t, deps[0].Outputs, []string{"connection_name"})
	assert.Equal(t, deps[1].Identity, api.ResourceIdentity{Type: "cloudsql", ID: "my-pg-db"})

	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `db_name = "my-pg-db"`)
	// assertInFile(t, filepath.Join(tmpDir, "cloudsql", "cloudsql-my-pg-db-dev", "main.tf"), `db_type = "POSTGRES_13"`)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `machine_type = "db-f1-micro"`)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `disk_size = 10`)
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `deletion_protection = true`)

}

func Test_ConfigureResource(t *testing.T) {
	db := &cloudSql{
		DbName: "theDb",
	}
	cloudRun := cloudRunConfig{}

	err := db.ConfigureResource(&cloudRun)
	assert.NoError(t, err)

	assert.Equal(t, cloudRun.DependsOn, []string{"module.cloudsql-theDb", "module.secret-cloudsql-theDb_PASSWORD", "module.secret-cloudsql-theDb_USER"})
	assert.Equal(t, cloudRun.Env.Secrets, map[string]string{
		"DB_theDb_PASSWORD": "module.secret-cloudsql-theDb_PASSWORD.secret_id",
		"DB_theDb_USER":     "module.secret-cloudsql-theDb_USER.secret_id",
	})
	assert.Equal(t, cloudRun.Env.Refs, map[string]string{
		"DB_theDb_HOST": "module.cloudsql-theDb.master_private_ip",
	})
}
