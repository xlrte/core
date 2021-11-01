package terraform

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Init_Terraform(t *testing.T) {
	tf, err := Init(".", os.Stdin, os.Stderr)
	assert.NoError(t, err)
	v, _, err := tf.Version(context.Background(), false)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}
