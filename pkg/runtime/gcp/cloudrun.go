package gcp

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

//go:embed templates/cloudrun.tf
var cloudRunMain string

//go:embed templates/cloudrun_domain.tf
var cloudRunNetworkMain string

type scaling struct {
	MinInstances int `yaml:"min_instances,omitempty"`
	MaxInstances int `yaml:"max_instances,omitempty"`
}
type cloudRunRuntimeConfig struct {
	Name        string  `yaml:"name" validate:"required"`
	Memory      string  `yaml:"memory,omitempty"`
	CPU         int     `yaml:"cpu,omitempty" validate:"min=1,max=4"` // 1, 2, 4
	Timeout     int     `yaml:"timeout,omitempty"`
	MaxRequests int     `yaml:"max_requests,omitempty"`
	Scaling     scaling `yaml:"scaling,omitempty"`
	Domain      domain  `yaml:"domain,omitempty"`
}

type domain struct {
	Name    string `yaml:"name,omitempty"`
	DNSZone string `yaml:"dns_zone,omitempty"`
}

type cloudRunNetwork struct {
	Domain  domain
	DNSName string // domain + "."
}

type cloudRunConfig struct {
	baseDir               string
	ServiceName           string
	ImageID               string
	Traffic               int
	IsPublic              bool
	Http2                 bool
	Env                   api.EnvVars
	RuntimeConfig         cloudRunRuntimeConfig
	NetworkConfig         *cloudRunNetwork
	PublishTopics         []string
	SubscribeTopics       []*subscription
	CloudStorage          []*gcsIAM
	ServerlessNetworkLink string
	HasServerlessNetwork  bool
	DependsOn             []string
}

type cloudRunSpec struct {
	BaseName string `yaml:"base_name" validate:"required"`
	Http     http   `yaml:"http"`
}

type http struct {
	Public bool `yaml:"public" validate:"required"`
	Http2  bool `yaml:"http2"`
}

type crFile struct {
	name   string
	tmplte string
}

type cloudRunLoader struct {
	baseDir string
	service *api.Service
}

func (loader *cloudRunLoader) Name() string {
	return "cloudrun"
}

func (loader *cloudRunLoader) Load(ctx api.EnvContext, service *api.Service, deploymentContext api.DeploymentContext) (api.Resource, error) {
	config, err := loader.toCloudRunSettings(ctx, service, deploymentContext)
	if err != nil {
		return nil, err
	}
	config.baseDir = loader.baseDir
	return config, nil
}

func (config *cloudRunConfig) Configure() error {
	return configureCloudRun(config.baseDir, *config)
}

func (config *cloudRunConfig) Identity() api.ResourceIdentity {
	return api.ResourceIdentity{Type: "cloudrun", ID: config.ServiceName}
}

func (loader *cloudRunLoader) toCloudRunSettings(ctx api.EnvContext, service *api.Service, deploymentContext api.DeploymentContext) (*cloudRunConfig, error) {
	serviceSettings := defaultCloudRunRuntimeConfig()
	if deploymentContext.Resources != nil {
		runtimeSettings, err := parseCloudRunRTEConfig(deploymentContext.Resources)
		if err != nil {
			return nil, err
		}
		for _, settings := range runtimeSettings {
			if settings.Name == service.Name() {
				serviceSettings = settings
				break
			}
		}
	}
	if ctx.RepoBase == "" {
		ctx.RepoBase = fmt.Sprintf("gcr.io/%s/", ctx.Context)
	}

	bytes, err := yaml.Marshal(service.Spec)
	if err != nil {
		return nil, err
	}
	var def cloudRunSpec
	err = yaml.Unmarshal(bytes, &def)
	if err != nil {
		return nil, err
	}
	version, err := ctx.Version(service.SVCName)
	if err != nil {
		return nil, err
	}
	config := &cloudRunConfig{
		ServiceName:   service.SVCName,
		ImageID:       fmt.Sprintf("%s%s:%s", ctx.RepoBase, def.BaseName, version),
		Traffic:       100,
		IsPublic:      def.Http.Public,
		Http2:         def.Http.Http2,
		RuntimeConfig: *serviceSettings,
		Env:           deploymentContext.Env,
	}
	if config.Env.Refs == nil {
		config.Env.Refs = make(map[string]string)
	}
	if config.Env.Secrets == nil {
		config.Env.Secrets = make(map[string]string)
	}

	for k, v := range config.Env.Secrets {
		config.Env.Secrets[k] = fmt.Sprintf("module.secret-%s.secret_id", v)
		config.DependsOn = append(config.DependsOn, fmt.Sprintf("module.secret-%s", v))
	}

	if config.RuntimeConfig.Domain.DNSZone != "" && config.RuntimeConfig.Domain.Name != "" {

		nwConfig := cloudRunNetwork{
			Domain:  config.RuntimeConfig.Domain,
			DNSName: fmt.Sprintf("%s.", config.RuntimeConfig.Domain.Name),
		}
		config.NetworkConfig = &nwConfig

	}

	return config, nil
}

func defaultCloudRunRuntimeConfig() *cloudRunRuntimeConfig {
	return &cloudRunRuntimeConfig{
		Memory:      "512Mi",
		CPU:         1,
		Timeout:     300,
		MaxRequests: 80,
		Scaling: scaling{
			MinInstances: 0,
			MaxInstances: 100,
		},
	}
}

func parseCloudRunRTEConfig(resource *[]byte) ([]*cloudRunRuntimeConfig, error) {
	var configs = []*cloudRunRuntimeConfig{}
	err := yaml.Unmarshal(*resource, &configs)
	if err != nil {
		return nil, err
	}

	for _, conf := range configs {
		if conf.CPU == 0 {
			conf.CPU = 1
		}
		if conf.MaxRequests == 0 {
			conf.MaxRequests = 80
		}
		if conf.Memory == "" {
			conf.Memory = "512Mi"
		}
		if conf.Timeout == 0 {
			conf.Timeout = 300
		}
		if conf.Scaling.MaxInstances == 0 {
			conf.Scaling.MaxInstances = 1000
		}
		validate := validator.New()
		if errs := validate.Struct(conf); errs != nil {
			return nil, errs
		}
	}

	return configs, nil
}

func configureCloudRun(baseDir string, config cloudRunConfig) error {
	files := []crFile{
		{"cloudrun.tf", cloudRunMain},
	}

	err := applyTerraformTemplates(baseDir, files, config)
	if err != nil {
		return err
	}
	if config.NetworkConfig != nil {
		networkFiles := []crFile{
			{"dns.tf", cloudRunNetworkMain},
		}
		return applyTerraformTemplates(baseDir, networkFiles, config)
	}
	return nil
}

func applyTerraformTemplates(baseDir string, files []crFile, config interface{}) error {
	mainFile := filepath.Join(baseDir, "main.tf")
	data, err := ioutil.ReadFile(filepath.Clean(mainFile))
	if err != nil {
		_, err = os.Create(filepath.Clean(mainFile))
		if err != nil {
			return err
		}
		data = []byte{}
	}

	for _, file := range files {
		tmpl, e := template.New(file.name).Parse(file.tmplte)
		if e != nil {
			return e
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, config)
		if err != nil {
			return err
		}

		output := buf.Bytes()
		data = append(data, output...)
	}
	err = os.WriteFile(mainFile, data, 0600)

	return err
}
