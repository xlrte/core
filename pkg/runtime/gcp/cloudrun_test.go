package gcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_Template(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	conf := cloudRunConfig{
		ServiceName:   "test-srv",
		ImageID:       "gcr.io/exlrte/test-srv:foo",
		Traffic:       100,
		IsPublic:      true,
		PublishTopics: []string{"a-topic"},
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "100")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "xlrte")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "gcr.io/exlrte/test-srv:foo")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "test-srv")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "a-topic")

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "cloud_run_endpoint")
}

func Test_Template_SubscribeTopics(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	conf := cloudRunConfig{
		ServiceName: "test-srv",
		ImageID:     "gcr.io/exlrte/test-srv:foo",
		Traffic:     100,
		IsPublic:    true,
		SubscribeTopics: []*subscription{
			{
				TopicName:             "foo-topic",
				AckDeadline:           33,
				Retention:             "605s",
				EnableMessageOrdering: true,
				RetainAckedMessages:   true,
			},
		},
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "foo-topic")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "ack_deadline_seconds = 33")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), `message_retention_duration = "605s"`)
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "retain_acked_messages = true")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "enable_message_ordering = true")
}

func Test_Template_With_Network(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	conf := cloudRunConfig{
		ServiceName: "test-srv",
		ImageID:     "gcr.io/exlrte/test-srv:foo",
		Traffic:     100,
		IsPublic:    true,
		NetworkConfig: &cloudRunNetwork{
			Domain: domain{
				Name:    "xlrte.org",
				DNSZone: "dnsZone",
			},
			DNSName: "xlrte.com.",
		},
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "xlrte.org")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "dnsZone")
	assertInFile(t, filepath.Join(tmpDir, "main.tf"), "xlrte.com.")

}

func Test_Exclusion_From_Template(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer func() {
		e := os.RemoveAll(tmpDir)
		assert.NoError(t, e)
	}()

	conf := cloudRunConfig{
		ServiceName: "test-srv",
		ImageID:     "gcr.io/exlrte/test-srv:foo",
		Traffic:     100,
		IsPublic:    false,
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "main.tf")) // nolint
	assert.NoError(t, err)
	str := string(b)
	assert.Contains(t, str, "http2 = false")
	// assert.Contains(t, str, `default = "http1"`)
}

func Test_ReadCloudRun_Config(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-config.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	configs, err := parseCloudRunRTEConfig(&bytes)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(configs))

	assert.Equal(t, 2, configs[0].CPU)
	assert.Equal(t, "1024Mi", configs[0].Memory)
	assert.Equal(t, 10, configs[0].MaxRequests)
	assert.Equal(t, 100, configs[0].Timeout)
	assert.Equal(t, 10, configs[0].Scaling.MaxInstances)
	assert.Equal(t, 1, configs[0].Scaling.MinInstances)
	assert.Equal(t, "cloudrun-srv", configs[0].Name)

}

func assertInFile(t *testing.T, file, assertion string) {
	b, err := ioutil.ReadFile(file) // nolint
	assert.NoError(t, err)
	str := string(b)
	assert.True(t, strings.Contains(str, assertion), fmt.Sprintf("The string %s does not contain the assertion %s", str, assertion))

}

func Test_parseCloudRunSettings_No_CPU(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-nocpu.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	conf, err := parseCloudRunRTEConfig(&bytes)
	assert.NoError(t, err)

	assert.Equal(t, 1, conf[0].CPU)
	assert.Equal(t, "512Mi", conf[0].Memory)
	assert.Equal(t, 80, conf[0].MaxRequests)
	assert.Equal(t, 300, conf[0].Timeout)
	assert.Equal(t, 1000, conf[0].Scaling.MaxInstances)
	assert.Equal(t, 0, conf[0].Scaling.MinInstances)
	assert.Equal(t, "cloudrun-srv", conf[0].Name)
}

func Test_parseCloudRunSettings_HasMissConf(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Clean(filepath.Join("testdata", "cloudrun", "cloudrun-missconfigured.yaml")))
	assert.NoError(t, err)
	var theMap map[string]interface{}
	err = yaml.Unmarshal(data, &theMap)
	assert.NoError(t, err)
	assert.NotNil(t, theMap["cloudrun"])
	bytes, err := yaml.Marshal(theMap["cloudrun"])
	assert.NoError(t, err)

	_, err = parseCloudRunRTEConfig(&bytes)
	assert.Error(t, err)

}
