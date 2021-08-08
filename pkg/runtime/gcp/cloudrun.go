package gcp

import (
	_ "embed"
	"html/template"
	"os"
	"path/filepath"

	"github.com/xlrte/core/pkg/api"
)

//go:embed cloudrun/variables.tf
var cloudRunVars string

//go:embed cloudrun/outputs.tf
var cloudRunOutputs string

//go:embed cloudrun/main.tf
var cloudRunMain string

type cloudRunConfig struct {
	ServiceName string
	ImageID     string
	Region      string
	Project     string
	Traffic     int
	IsPublic    bool
	Env         *api.EnvVars
}

type crFile struct {
	name   string
	tmplte string
}

func configureCloudRun(baseDir string, config cloudRunConfig) error {
	files := []crFile{
		{"variables.tf", cloudRunVars},
		{"outputs.tf", cloudRunOutputs},
		{"main.tf", cloudRunMain},
	}

	dir := filepath.Join(baseDir, "services", config.ServiceName)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	for _, file := range files {
		tmpl, err := template.New(file.name).Parse(file.tmplte)
		if err != nil {
			return err
		}
		outputFile := filepath.Join(dir, file.name)
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		err = tmpl.Execute(f, config)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
