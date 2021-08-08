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
)

func Test_Template(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	conf := cloudRunConfig{
		ServiceName: "test-srv",
		ImageID:     "gcr.io/exlrte/test-srv:foo",
		Region:      "europe-west6",
		Project:     "xlrte",
		Traffic:     100,
		IsPublic:    true,
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "variables.tf"), "default = 100")
	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "variables.tf"), "xlrte")
	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "variables.tf"), "europe-west6")
	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "variables.tf"), "gcr.io/exlrte/test-srv:foo")
	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "variables.tf"), "test-srv")

	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "outputs.tf"), "cloud_run_endpoint")
	assertInFile(t, filepath.Join(tmpDir, "services", "test-srv", "main.tf"), "google_iam_policy")
}

func Test_Exclusion_From_Template(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tf_temp")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	conf := cloudRunConfig{
		ServiceName: "test-srv",
		ImageID:     "gcr.io/exlrte/test-srv:foo",
		Region:      "europe-west6",
		Project:     "xlrte",
		Traffic:     100,
		IsPublic:    false,
	}

	err = configureCloudRun(tmpDir, conf)
	assert.NoError(t, err)

	assertNotInFile(t, filepath.Join(tmpDir, "services", "test-srv", "main.tf"), "google_iam_policy")
}

func assertInFile(t *testing.T, file, assertion string) {
	b, err := ioutil.ReadFile(file) // just pass the file name
	assert.NoError(t, err)
	str := string(b)
	assert.True(t, strings.Contains(str, assertion), fmt.Sprintf("The string %s does not contain the assertion %s", str, assertion))
}

func assertNotInFile(t *testing.T, file, assertion string) {
	b, err := ioutil.ReadFile(file) // just pass the file name
	assert.NoError(t, err)
	str := string(b)
	assert.False(t, strings.Contains(str, assertion), fmt.Sprintf("The string %s does not contain the assertion %s", str, assertion))
}
