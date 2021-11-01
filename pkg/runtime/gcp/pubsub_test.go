package gcp

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xlrte/core/pkg/api"
)

func Test_PubSub_Can_Read_Resources(t *testing.T) {
	serviceData := getCloudRunBytes(t, filepath.Join("testdata", "pubsub", "service.yaml"), "pubsub")
	confData := getCloudRunBytes(t, filepath.Join("testdata", "pubsub", "resources.yaml"), "pubsub")

	resource := &pubSubConfig{}
	resources, bindings, err := resource.Load(&api.ResourceDefinition{
		Name:           "pubsub",
		DependedOnBy:   api.ResourceIdentity{ID: "the-service", Type: "cloudrun"},
		ServiceConfig:  serviceData,
		ResourceConfig: &confData,
	})

	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Len(t, bindings, 3)

	assert.Equal(t, resources[0].Identity(), api.ResourceIdentity{Type: "pubsub", ID: "third_type_of_topic"})
	assert.Equal(t, resources[1].Identity(), api.ResourceIdentity{Type: "pubsub", ID: "some_other_topic"})

	assert.Equal(t, bindings[0].DependedOnBy, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"})
	assert.Equal(t, bindings[0].Identity, api.ResourceIdentity{Type: "pubsub", ID: "third_type_of_topic"})
	assert.Equal(t, bindings[0].Privileges, api.Owner)
	assert.Equal(t, *bindings[0].Config.(*publishDestination), publishDestination{"third_type_of_topic"})

	assert.Equal(t, bindings[1].DependedOnBy, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"})
	assert.Equal(t, bindings[1].Identity, api.ResourceIdentity{Type: "pubsub", ID: "some_topic"})
	assert.Equal(t, bindings[1].Privileges, api.ReadOnly)
	assert.Equal(t, *bindings[1].Config.(*subscription), subscription{
		TopicName:             "some_topic",
		AckDeadline:           30,
		Retention:             "300000s",
		EnableMessageOrdering: true,
		RetainAckedMessages:   true,
	})
	conf := defaultConf()
	conf.TopicName = "some_other_topic"
	assert.Equal(t, bindings[2].DependedOnBy, api.ResourceIdentity{ID: "the-service", Type: "cloudrun"})
	assert.Equal(t, bindings[2].Identity, api.ResourceIdentity{Type: "pubsub", ID: "some_other_topic"})
	assert.Equal(t, bindings[2].Privileges, api.Owner)
	assert.Equal(t, bindings[2].Config, conf)

}

func Test_PubSubTemplate(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	serviceData := getCloudRunBytes(t, filepath.Join("testdata", "pubsub", "service.yaml"), "pubsub")

	resource := &pubSubConfig{baseDir: tmpDir}
	resources, _, err := resource.Load(&api.ResourceDefinition{
		Name:           "pubsub",
		DependedOnBy:   api.ResourceIdentity{ID: "the-service", Type: "cloudrun"},
		ServiceConfig:  serviceData,
		ResourceConfig: nil,
	})
	assert.NoError(t, err)
	err = resources[0].Configure()

	assert.NoError(t, err)
	err = resources[1].Configure()

	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `"third_type_of_topic"`)
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `"some_other_topic"`)
}

func Test_ConfigureResource_Subs(t *testing.T) {
	resource := defaultConf()
	resource.TopicName = "the_topic"
	cloudRun := cloudRunConfig{}

	err := resource.ConfigureResource(&cloudRun)
	assert.NoError(t, err)
	assert.Equal(t, cloudRun.DependsOn, []string{"module.pubsub-the_topic.topic"})

	assert.Equal(t, cloudRun.SubscribeTopics, []*subscription{resource})
}

func Test_ConfigureResource_Publish(t *testing.T) {
	resource := &publishDestination{"the_topic"}
	cloudRun := cloudRunConfig{}

	err := resource.ConfigureResource(&cloudRun)
	assert.NoError(t, err)

	assert.Equal(t, cloudRun.DependsOn, []string{"module.pubsub-the_topic.topic"})

	assert.Equal(t, cloudRun.PublishTopics, []string{"the_topic"})
}
