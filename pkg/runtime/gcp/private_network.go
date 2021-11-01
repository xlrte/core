package gcp

import (
	"fmt"

	_ "embed"

	"github.com/xlrte/core/pkg/api"
)

//go:embed templates/private_network.tf
var privateNetworkTF string

type privateNetwork struct {
	baseDir      string
	MinInstances int    `yaml:"min_instances"` // min 2
	MaxInstances int    `yaml:"max_instances"` // min 3, max 10
	InstanceType string `yaml:"instance_type"` // f1-micro, e2-standard-4
}

type privateNetworkBinding struct {
	identity api.ResourceIdentity
}

func (r *privateNetwork) Configure() error {
	return applyTerraformTemplates(r.baseDir, []crFile{
		{"private_network.tf", privateNetworkTF},
	}, r)
}

func (r *privateNetwork) Identity() api.ResourceIdentity {
	return api.ResourceIdentity{Type: "private_network", ID: "network"}
}

func (r *privateNetwork) configurator() api.DependencyVisitor {
	return &privateNetworkBinding{r.Identity()}
}

func (r *privateNetworkBinding) ConfigureResource(resource api.Resource) error {
	cloudsql, ok := resource.(*cloudSql)
	if ok {
		cloudsql.NetworkLink = fmt.Sprintf("module.%s-%s.network_self_link", r.identity.Type, r.identity.ID)
	}
	cloudrun, ok := resource.(*cloudRunConfig)
	if ok {
		serverlessConnector := fmt.Sprintf("module.%s-%s.serverless_connector", r.identity.Type, r.identity.ID)
		cloudrun.ServerlessNetworkLink = serverlessConnector
		cloudrun.HasServerlessNetwork = true
		cloudrun.DependsOn = append(cloudrun.DependsOn, serverlessConnector)
	}
	return nil
}
