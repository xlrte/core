package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api/secrets"
	"gopkg.in/yaml.v2"
)

var selector = ArgEnvResolver{
	EnvName:    "prod",
	ArgVersion: "v1",
}

var privateKey string

func init() {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "private-key.asc"))
	if err != nil {
		panic(err)
	}
	privateKey = string(data)

}

type dummyRuntime struct {
	ResourceTypes     []string
	serviceConfigured int
	resourceLoaders   []ResourceLoader
	secretsInited     bool
	secretsInServices map[string]string
}

type dummyResource struct {
	resourceName string
}

type cloudSql struct {
	Name       string `yaml:"name"`
	DBType     string `yaml:"type"`
	Size       string `yaml:"size"`
	isCorrect  bool
	isMigrated bool
}
type cloudSqlConfig struct {
	ResourceIdentity ResourceIdentity
}

type dummyCloudRun struct {
	Name         string
	Dependencies []ResourceIdentity
}

func Test_Random_String(t *testing.T) {
	for i := 0; i < 100; i++ {
		str := randStringRunes(i + 1)
		assert.Len(t, str, i+1)
		assert.False(t, strings.Contains(str, " "))
	}
}

func Test_No_Matching_Env_Found(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{}},
	}

	_, err := parseDeploymentConfig("testdata", &ArgEnvResolver{"foo", "v1"}, &runtimes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find a target environment for")
}

func Test_No_Services_Found(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{
			ResourceTypes: []string{"cloudsql"},
		}},
	}

	_, err := parseDeploymentConfig(filepath.Join("testdata", "noservices"), &selector, &runtimes)
	assert.Error(t, err)
}

func Test_Runtime_Does_Not_Support_Resource(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{
			ResourceTypes: []string{},
		}},
	}

	_, err := parseDeploymentConfig(filepath.Join("testdata", "valid-env"), &selector, &runtimes)
	assert.Error(t, err)
}

func Test_Valid_Environment(t *testing.T) {
	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{
			ResourceTypes: []string{"cloudsql", "pubsub", "gcs"},
		}},
	}

	defs, err := parseDeploymentConfig(filepath.Join("testdata", "valid-env"), &selector, &runtimes)
	assert.NoError(t, err)
	assert.Len(t, defs, 1)

	assert.Equal(t, defs[0].Runtime, runtimes.Runtimes[0])
	assert.Equal(t, defs[0].Environment.Name(), "prod")
	assert.Len(t, defs[0].Resources, 6)
	assert.Len(t, defs[0].Services, 2)
	assert.Len(t, defs[0].Services[0].Env.Vars, 1)
	assert.Len(t, defs[0].Services[1].Env.Vars, 1)

}

func Test_Valid_Deploy(t *testing.T) {
	secretsDir := filepath.Join("testdata", "valid-env", "environments", "prod", "secrets")
	err := os.RemoveAll(secretsDir)
	assert.NoError(t, err)
	err = os.MkdirAll(secretsDir, 0750)
	assert.NoError(t, err)
	err = os.Setenv("XLRTE_PRIVATE_KEY", privateKey)
	assert.NoError(t, err)
	err = os.Setenv("XLRTE_PASSPHRASE", "pass")
	assert.NoError(t, err)
	defer func() {
		err = os.Setenv("XLRTE_PRIVATE_KEY", "")
		assert.NoError(t, err)
		err = os.Setenv("XLRTE_PASSPHRASE", "")
		assert.NoError(t, err)
	}()

	runtimes := Runtimes{
		Runtimes: []Runtime{&dummyRuntime{
			ResourceTypes: []string{"cloudsql", "pubsub", "gcs"},
		}},
	}

	configs, preApply, err := Prepare(filepath.Join("testdata", "valid-env"), &selector, &runtimes)
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.NotNil(t, preApply)
	assert.True(t, runtimes.Runtimes[0].(*dummyRuntime).secretsInited)
	assert.NoError(t, preApply(context.Background()))
	assert.Greater(t, len(configs), 0)
	for _, conf := range configs {
		version, e := conf.Environment.Resolver.Version("foo")
		assert.NoError(t, e)
		assert.Equal(t, len(conf.underlyingResources), 4)
		assert.Equal(t, "v1", version)
		dRte := conf.Runtime.(*dummyRuntime)
		assert.Equal(t, 2, dRte.serviceConfigured)
		assert.Equal(t, map[string]string{
			"verySecret":  "here",
			"otherSecret": "theSecret",
		}, dRte.secretsInServices)
		for _, loader := range dRte.resourceLoaders {
			theLoader := loader.(*dummyResource)
			fmt.Println(theLoader)
		}
		i := 0
		for _, resource := range conf.underlyingResources {
			cloudsql, ok := resource.(*cloudSql)
			if ok {
				assert.True(t, cloudsql.isCorrect)
				assert.True(t, cloudsql.isMigrated)
				i++
			}
			cloudrun, ok := resource.(*dummyCloudRun)
			if ok {
				ri := ResourceIdentity{ID: "my-pg-db", Type: "cloudsql"}
				ri2 := ResourceIdentity{ID: "another-db", Type: "cloudsql"}
				isExpectedService := cloudrun.Name == "cloudrun-srv2" || cloudrun.Name == "cloudrun-srv"
				assert.True(t, isExpectedService)
				if cloudrun.Name == "cloudrun-srv2" {
					assert.Equal(t, cloudrun.Dependencies, []ResourceIdentity{ri})
				} else {
					assert.Equal(t, cloudrun.Dependencies, []ResourceIdentity{ri, ri2})
				}
			}
		}
		assert.Equal(t, i, 2, "Should only have one cloud sql instance despite being defined in two services")
	}
	secretFiles := 0
	err = filepath.Walk(secretsDir, func(path string, info os.FileInfo, e error) error {
		if strings.HasSuffix(path, ".asc") {
			secretFiles++
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, secretFiles)
}

func (rt *dummyRuntime) Name() string {
	return "cloudrun"
}

func (rt *dummyRuntime) Init(ctx EnvContext) error {
	return nil
}

func (rt *dummyRuntime) InitEnvironment(ctx context.Context, env, project, region string) error {
	return nil
}

func (rt *dummyRuntime) Resources() []ResourceLoader {
	if len(rt.resourceLoaders) > 0 {
		return rt.resourceLoaders
	}
	loaders := []ResourceLoader{}
	for _, resource := range rt.ResourceTypes {
		loaders = append(loaders, &dummyResource{resourceName: resource})
	}
	return loaders
}

func (rt *dummyRuntime) Services() []ServiceLoader {
	return []ServiceLoader{
		rt,
	}
}

func (rt *dummyRuntime) InitSecrets(env EnvContext, secrets []*secrets.Secret) error {
	rt.secretsInited = true
	return nil
}

func (rt *dummyRuntime) Load(ctx EnvContext, artifact *Service, deploymentContext DeploymentContext) (Resource, error) {

	if deploymentContext.Env.Vars["foo"] != "baz" {
		return nil, fmt.Errorf("environment is invalid: %v", deploymentContext.Env.Vars)
	}
	if deploymentContext.Env.Vars["bar"] != "" {
		return nil, fmt.Errorf("environment is invalid: %v", deploymentContext.Env.Vars)
	}
	if deploymentContext.Resources == nil {
		panic("resources should have been set")
	}
	if len(deploymentContext.Env.Secrets) != 1 {
		panic(artifact.SVCName + " should only have one secret")
	}
	for k, v := range deploymentContext.Env.Secrets {
		if rt.secretsInServices == nil {
			rt.secretsInServices = map[string]string{}
		}
		_, found := rt.secretsInServices[k]
		if found {
			panic("should not have duplicate of secret " + k)
		}
		rt.secretsInServices[k] = v
	}

	rt.serviceConfigured = rt.serviceConfigured + 1
	return &dummyCloudRun{Name: artifact.SVCName}, nil
}

func (cr *dummyCloudRun) Identity() ResourceIdentity {
	return ResourceIdentity{Type: "cloudrun", ID: cr.Name}
}

func (cr *dummyCloudRun) Configure() error {
	return nil
}

func (rt *dummyRuntime) Apply(ctx context.Context) error {
	return nil
}
func (rt *dummyRuntime) Plan(ctx context.Context) error {
	return nil
}
func (rt *dummyRuntime) Delete(ctx context.Context) error {
	return nil
}
func (rt *dummyRuntime) Export(ctx context.Context) error {
	return nil
}

func (rt *dummyResource) Name() string {
	return rt.resourceName
}

func (rt *dummyResource) Identity() ResourceIdentity {
	return ResourceIdentity{rt.resourceName, rt.Name()}
}

func (rt *dummyResource) Load(d *ResourceDefinition) ([]Resource, []DependencyBinding, error) {
	var dbs []*cloudSql
	var resources []cloudSql
	var rs []Resource

	if d.Name == "cloudsql" {
		err := yaml.Unmarshal(d.ServiceConfig, &dbs)
		if err != nil {
			return nil, nil, err
		}
		err = yaml.Unmarshal(*d.ResourceConfig, &resources)
		if err != nil {
			return nil, nil, err
		}
	}
	var bindings []DependencyBinding
	for _, db := range dbs {
		for _, r := range resources {
			if db.Name == r.Name {
				db.Size = r.Size
			}
		}
		bindings = append(bindings, DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   Owner,
			Identity:     ResourceIdentity{Type: "cloudsql", ID: db.Name},
			Config:       &cloudSqlConfig{ResourceIdentity{Type: "cloudsql", ID: db.Name}},
			SecretRefs:   []SecretRef{{Name: "USERNAME", Type: RandomString}},
		})
		rs = append(rs, db)
	}

	return rs, bindings, nil
}

func (r *cloudSql) Configure() error {
	if r.DBType != "" && r.Name != "" && r.Size != "" {
		r.isCorrect = true
	}
	return nil
}

func (r *cloudSqlConfig) ConfigureResource(resource Resource) error {
	cr, ok := resource.(*dummyCloudRun)
	if ok {
		cr.Dependencies = append(cr.Dependencies, r.ResourceIdentity)
	}
	return nil
}

func (r *cloudSql) Migrate(ctx context.Context) error {
	r.isMigrated = true
	return nil
}

func (r *cloudSql) Identity() ResourceIdentity {
	return ResourceIdentity{Type: "cloudsql", ID: r.Name}
}
