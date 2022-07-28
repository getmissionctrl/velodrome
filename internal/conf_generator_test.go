package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTerraform(t *testing.T) {
	config, err := LoadConfig("testdata/config.yaml")
	assert.NoError(t, err)
	err = GenerateTerraform(config)
	assert.NoError(t, err)
	assert.True(t, false)
}
