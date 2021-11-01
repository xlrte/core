package gcp

import (
	_ "embed"
	"fmt"

	"github.com/xlrte/core/pkg/api"
	"gopkg.in/yaml.v2"
)

// We need pubsub creation
// service account,
//go:embed templates/pubsub.tf
var pubsubMain string

type pubSubConfig struct {
	baseDir               string
	TopicName             string `yaml:"name"`
	Owner                 bool   `yaml:"owner"`
	AckDeadline           int    `yaml:"ack_deadline_seconds"`
	Retention             int    `yaml:"message_retention_duration"`
	EnableMessageOrdering bool   `yaml:"enable_message_ordering"`
	RetainAckedMessages   bool   `yaml:"retain_acked_messages"`
}

type subscription struct {
	TopicName             string
	AckDeadline           int
	Retention             string
	EnableMessageOrdering bool
	RetainAckedMessages   bool
}

type publishDestination struct {
	TopicName string
}

func defaultConf() *subscription {
	return &subscription{
		TopicName:             "",
		AckDeadline:           20,
		Retention:             "604800s",
		EnableMessageOrdering: false,
		RetainAckedMessages:   false,
	}
}

func (rt *pubSubConfig) toConfig() *subscription {
	theMap := defaultConf()
	if rt.AckDeadline > 0 {
		theMap.AckDeadline = rt.AckDeadline
	}

	if rt.Retention > 0 {
		theMap.Retention = fmt.Sprintf("%ds", rt.Retention)
	}

	theMap.EnableMessageOrdering = rt.EnableMessageOrdering
	theMap.RetainAckedMessages = rt.RetainAckedMessages
	theMap.TopicName = rt.TopicName
	return theMap
}

func toConfig(id api.ResourceIdentity, configs []pubSubConfig) *subscription {
	for _, pbc := range configs {
		if pbc.Identity() == id {
			return pbc.toConfig()
		}
	}
	conf := defaultConf()
	conf.TopicName = id.ID
	return conf
}

func (rt *pubSubConfig) Load(d *api.ResourceDefinition) ([]api.Resource, []api.DependencyBinding, error) {
	var rs []api.Resource
	var bindings []api.DependencyBinding

	var settings map[string][]pubSubConfig
	var resources []pubSubConfig

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

	for index := range settings["produce"] {
		settings["produce"][index].baseDir = rt.baseDir
		rs = append(rs, &settings["produce"][index])
		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   api.Owner,
			Identity:     settings["produce"][index].Identity(),
			Config:       &publishDestination{settings["produce"][index].Identity().ID},
		})
	}
	for index := range settings["consume"] {
		privilege := api.ReadOnly
		if settings["consume"][index].Owner {
			settings["consume"][index].baseDir = rt.baseDir
			rs = append(rs, &settings["consume"][index])
			privilege = api.Owner
		}
		bindings = append(bindings, api.DependencyBinding{
			DependedOnBy: d.DependedOnBy,
			Privileges:   privilege,
			Identity:     settings["consume"][index].Identity(),
			Config:       toConfig(settings["consume"][index].Identity(), resources),
		})
	}

	return rs, bindings, nil
}

func (r *pubSubConfig) Configure() error {
	return applyTerraformTemplates(r.baseDir, []crFile{
		{"pubsub.tf", pubsubMain},
	}, r)
}

func (r *pubSubConfig) Name() string {
	return "pubsub"
}

func (r *pubSubConfig) Identity() api.ResourceIdentity {
	return api.ResourceIdentity{Type: "pubsub", ID: r.TopicName}
}

func (sub *subscription) ConfigureResource(resource api.Resource) error {
	crConfig, ok := resource.(*cloudRunConfig)
	if ok {
		topic := fmt.Sprintf("module.%s-%s.topic", "pubsub", sub.TopicName)
		crConfig.DependsOn = append(crConfig.DependsOn, topic)
		crConfig.SubscribeTopics = append(crConfig.SubscribeTopics, sub)
	}
	return nil
}

func (pub *publishDestination) ConfigureResource(resource api.Resource) error {
	crConfig, ok := resource.(*cloudRunConfig)
	if ok {
		topic := fmt.Sprintf("module.%s-%s.topic", "pubsub", pub.TopicName)
		crConfig.DependsOn = append(crConfig.DependsOn, topic)
		crConfig.PublishTopics = append(crConfig.PublishTopics, pub.TopicName)
	}
	return nil
}
