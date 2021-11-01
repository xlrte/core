package gcp

import (
	_ "embed"
	"fmt"

	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

//go:embed templates/cloudstorage.tf
var gcsMain string

type gcsConfig struct {
	baseDir           string
	BucketName        string `yaml:"name"`
	IsPublic          bool   `yaml:"public"`
	Access            string `yaml:"access"`
	Owner             *bool  `yaml:"owner"`
	Location          string `yaml:"location"`
	StorageClass      string `yaml:"storage_class"`
	VersioningEnabled *bool  `yaml:"versioning_enabled"`
}

type gcsIAM struct {
	Bucket string
	Role   string
}

func (rt *gcsConfig) Load(d *api.ResourceDefinition) ([]api.Resource, []api.DependencyBinding, error) {
	var rs []api.Resource
	var bindings []api.DependencyBinding

	var settings []*gcsConfig
	var resources []*gcsConfig

	err := yaml.Unmarshal(d.ServiceConfig, &settings)
	if err != nil {
		return nil, nil, err
	}
	if d.ResourceConfig != nil {
		err = yaml.Unmarshal(*d.ResourceConfig, &resources)
		if err != nil {
			return nil, nil, err
		}
	}

	for _, dep := range settings {
		ownership := api.ReadOnly
		if dep.Owner != nil {
			if !*dep.Owner && dep.Access == "readwrite" {
				ownership = api.ReadWrite
			} else {
				ownership = api.Owner
			}
		} else if dep.Access == "readwrite" {
			ownership = api.Owner
		}
		dep.Location = "US"
		dep.StorageClass = "STANDARD"
		enabled := false
		dep.VersioningEnabled = &enabled

		for _, res := range resources {
			if res.Identity() == dep.Identity() {
				if res.Location != "" {
					dep.Location = res.Location
				}
				if res.StorageClass != "" {
					dep.StorageClass = res.StorageClass
				}
				if res.VersioningEnabled != nil {
					dep.VersioningEnabled = res.VersioningEnabled
				}
				break
			}
		}
		dep.baseDir = rt.baseDir
		iamRole := gcsIAM{dep.BucketName, "roles/storage.objectViewer"}
		if ownership == api.Owner || ownership == api.ReadWrite {
			iamRole = gcsIAM{dep.BucketName, "roles/storage.objectAdmin"}
		}

		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   ownership,
			Identity:     dep.Identity(),
			Config:       &iamRole,
		})
		if ownership == api.Owner {
			rs = append(rs, dep)
		}
	}

	return rs, bindings, nil
}

func (r *gcsConfig) Configure() error {
	return applyTerraformTemplates(r.baseDir, []crFile{
		{"main.tf", gcsMain},
	}, r)
}

func (r *gcsConfig) Name() string {
	return "cloudstorage"
}

func (r *gcsConfig) Identity() api.ResourceIdentity {
	return api.ResourceIdentity{Type: r.Name(), ID: r.BucketName}
}

func (iam *gcsIAM) ConfigureResource(resource api.Resource) error {
	cloudRun, ok := resource.(*cloudRunConfig)
	if ok {
		bucket := fmt.Sprintf("module.%s-%s.bucket", "cloudstorage", iam.Bucket)
		cloudRun.DependsOn = append(cloudRun.DependsOn, bucket)
		cloudRun.CloudStorage = append(cloudRun.CloudStorage, iam)
	}
	return nil
}
