package terraform

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
)

func Init(workingDir string, stdOut, stdErr io.Writer) (*tfexec.Terraform, error) {

	execPath, err := tfinstall.Find(context.Background(), &tfinstall.LookPathOption{})
	if err != nil {
		tmpDir, err2 := ioutil.TempDir("", "tfinstall")
		if err2 != nil {
			return nil, err2
		}
		defer func() {
			e := os.RemoveAll(filepath.Clean(tmpDir))
			if e != nil {
				panic(e)
			}
		}()
		execPath, err2 = tfinstall.Find(context.Background(), tfinstall.LatestVersion(tmpDir, false))
		if err2 != nil {
			return nil, err2
		}
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, err
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return nil, err
	}

	tf.SetStdout(stdOut)
	tf.SetStderr(stdErr)
	return tf, nil
}
