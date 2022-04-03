package terraform

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Init_Terraform(t *testing.T) {
	ctx := context.Background()
	tf, err := Init(ctx, ".", os.Stdin, os.Stderr)
	assert.NoError(t, err)
	v, _, err := tf.Version(ctx, false)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}
