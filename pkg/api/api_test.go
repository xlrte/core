package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FileResolver_Resolves_Version(t *testing.T) {
	selector := FileEnvResolver{EnvName: "dev", BaseDir: "testdata"}
	v, err := selector.Version("theService")
	assert.NoError(t, err)
	assert.Equal(t, "v2", v)

}

func Test_FileResolver_No_Version_Defined(t *testing.T) {
	selector := FileEnvResolver{EnvName: "dev", BaseDir: "testdata"}
	v, err := selector.Version("theService2")
	assert.Error(t, err)
	assert.Equal(t, "", v)
}

func Test_FileResolver_No_File_Defined(t *testing.T) {
	selector := FileEnvResolver{EnvName: "prod", BaseDir: "testdata"}
	v, err := selector.Version("theService")
	assert.Error(t, err)
	assert.Equal(t, "", v)
}
