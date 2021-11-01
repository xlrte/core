package gcp

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xlrte/core/pkg/api"
	"github.com/xlrte/core/pkg/api/secrets"
	"github.com/xlrte/core/pkg/terraform"
)

//go:embed modules/*
var modules embed.FS

//go:embed templates/main.tf
var runtimeMain string

//go:embed templates/init.tf
var initMain string

//go:embed templates/secret.tf
var secretMain string

type gcpRuntime struct {
	modulesDir  string
	baseDir     string
	Region      string
	Project     string
	Environment string
	resetVars   []string
}

func NewRuntime(modulesDir string, baseDir string) api.Runtime {
	mainFile := filepath.Join(baseDir, "main.tf")
	os.Remove(mainFile) //nolint

	return &gcpRuntime{modulesDir: modulesDir, baseDir: baseDir, resetVars: []string{}}
}

func (rt *gcpRuntime) InitEnvironment(ctx context.Context, env, project, region string) error {
	initDir := filepath.Join(rt.baseDir, "init")

	fmt.Println("`xlrte init` sets up a GCP project and enables the required GCP services.")
	fmt.Println("You should not run this against an existing project unless it was created with xlrte, as it may disrupt previous project settings.")
	fmt.Println("If you want to run xlrte against an existing GCP project, please refer to the documentation to see which services you need to enable.")
	fmt.Println("Are you sure you want to continue? ('yes', or any other input for no)")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if text != "yes\n" && text != "yes" {
		fmt.Println("Aborting project setup")
		return nil
	}

	err = os.MkdirAll(filepath.Clean(initDir), 0750)
	if err != nil {
		return err
	}
	err = os.RemoveAll(filepath.Join(initDir, "main.tf"))
	if err != nil {
		return err
	}
	billingAccount := os.Getenv("GCP_BILLING_ACCOUNT_ID")

	if billingAccount == "" {
		return fmt.Errorf(`missing environment variable for GCP_BILLING_ACCOUNT_ID. 
Please set this environment variable in order to be able to initialize a GCP project fully.
You can find this under "Billing Accounts" in cloud.google.com`)
	}

	config := struct {
		Project        string
		Region         string
		BillingAccount string
	}{
		project,
		region,
		billingAccount,
	}
	err = applyTerraformTemplates(initDir, []crFile{{
		"main.tf", initMain,
	}}, &config)
	if err != nil {
		return err
	}
	tf, err := terraform.Init(initDir, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return tf.Apply(ctx)
}

func (rt *gcpRuntime) Name() string {
	return "gcp"
}

func (rt *gcpRuntime) resetEnv() error {
	for _, varName := range rt.resetVars {
		err := os.Setenv(varName, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func (rt *gcpRuntime) InitSecrets(env api.EnvContext, secrets []*secrets.Secret) error {

	for _, secret := range secrets {
		key := fmt.Sprintf("TF_VAR_secret_%s", secret.Name)
		err := os.Setenv(key, secret.Value)
		rt.resetVars = append(rt.resetVars, key)
		if err != nil {
			return err
		}
		err = applyTerraformTemplates(rt.baseDir, []crFile{
			{"secret.tf", secretMain},
		}, secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *gcpRuntime) Services() []api.ServiceLoader {
	return []api.ServiceLoader{
		&cloudRunLoader{baseDir: rt.baseDir, service: nil},
	}
}
func (rt *gcpRuntime) Resources() []api.ResourceLoader {
	return []api.ResourceLoader{
		&cloudSql{baseDir: rt.baseDir},
		&pubSubConfig{baseDir: rt.baseDir},
		&gcsConfig{baseDir: rt.baseDir},
		&cloudRunDependency{baseDir: rt.baseDir},
	}
}

func (rt *gcpRuntime) Init(ctx api.EnvContext) error {
	rt.Project = ctx.Context
	rt.Region = ctx.Region
	rt.Environment = ctx.EnvName
	err := rt.setProvider()
	return err
}

func (rt *gcpRuntime) Apply(ctx context.Context) error {
	defer func() {
		_ = rt.resetEnv()
	}()
	return rt.execCommand(ctx, api.Apply)
}
func (rt *gcpRuntime) Plan(ctx context.Context) error {
	defer func() {
		_ = rt.resetEnv()
	}()
	return rt.execCommand(ctx, api.Plan)
}

func (rt *gcpRuntime) Delete(ctx context.Context) error {
	defer func() {
		_ = rt.resetEnv()
	}()
	return rt.execCommand(ctx, api.Delete)
}

func (rt *gcpRuntime) Export(ctx context.Context) error {
	defer func() {
		_ = rt.resetEnv()
	}()
	return rt.execCommand(ctx, api.Export)
}

func (rt *gcpRuntime) execCommand(ctx context.Context, cmd api.Command) error {
	tf, err := terraform.Init(rt.baseDir, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	switch cmd {
	case api.Plan:
		_, err = tf.Plan(ctx)
		return err
	case api.Export:
		return nil
	case api.Apply:
		return tf.Apply(ctx)
	case api.Delete:
		return tf.Destroy(ctx)
	}

	return nil
}

func copyModules(entries []fs.DirEntry, fsPath string, targetDir string) error {
	for _, e := range entries {
		currentPath := filepath.Join(fsPath, e.Name())
		toMake := filepath.Join(targetDir, currentPath)
		if e.IsDir() {
			err := os.MkdirAll(toMake, 0750)
			if err != nil {
				return err
			}
			files, err := modules.ReadDir(currentPath)
			if err != nil {
				return err
			}
			err = copyModules(files, currentPath, targetDir)
			if err != nil {
				return err
			}
		} else {
			bytes, err := fs.ReadFile(modules, currentPath)
			if err != nil {
				return err
			}
			f, err := os.Create(toMake)
			if err != nil {
				return err
			}
			defer f.Close() //nolint
			_, err = f.Write(bytes)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (rt *gcpRuntime) setProvider() error {
	dir, err := modules.ReadDir(".")
	if err != nil {
		return err
	}
	err = copyModules(dir, "", rt.modulesDir)
	if err != nil {
		return err
	}
	tmpl, err := template.New("main.tf").Parse(runtimeMain)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, rt)
	if err != nil {
		return err
	}
	mainFile := filepath.Join(rt.baseDir, "main.tf")
	data, err := ioutil.ReadFile(filepath.Clean(mainFile))
	if err != nil {
		_, err = os.Create(mainFile)
		if err != nil {
			return err
		}
		data = []byte{}
	}

	output := buf.Bytes()
	data = append(output, data...)
	err = os.WriteFile(mainFile, data, 0600)

	return err
}
