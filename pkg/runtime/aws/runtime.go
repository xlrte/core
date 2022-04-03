package aws

import (
	"context"
	"os"
	"path/filepath"

	"github.com/xlrte/core/pkg/api"
	"github.com/xlrte/core/pkg/api/secrets"
)

type awsRuntime struct {
}

func NewRuntime(modulesDir string, baseDir string) api.Runtime {
	mainFile := filepath.Join(baseDir, "main.tf")
	os.Remove(mainFile) //nolint

	return &awsRuntime{}
}

func (rt *awsRuntime) Name() string {
	return "aws"
}

// InitEnvironment initialises an environment for the first time, such as 'dev', 'prod' etc.
func (rt *awsRuntime) InitEnvironment(ctx context.Context, env, project, region string) error {
	return nil
}

//Init initialises for a plan or apply
func (rt *awsRuntime) Init(env api.EnvContext) error {
	return nil
}

//InitSecrets initialises the secrets system.
func (rt *awsRuntime) InitSecrets(env api.EnvContext, secrets []*secrets.Secret) error {
	return nil
}
func (rt *awsRuntime) Resources() []api.ResourceLoader {
	return nil
}
func (rt *awsRuntime) Services() []api.ServiceLoader {
	return nil
}
func (rt *awsRuntime) Apply(ctx context.Context) error {
	return nil
}
func (rt *awsRuntime) Plan(ctx context.Context) error {
	return nil
}
func (rt *awsRuntime) Delete(ctx context.Context) error {
	return nil
}
func (rt *awsRuntime) Export(ctx context.Context) error {
	return nil
}
