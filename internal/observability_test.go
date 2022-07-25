package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMkObservabilityConfigs(t *testing.T) {
	folder := RandString(8)
	os.MkdirAll(filepath.Join(folder, "secrets"), 0755)

	os.MkdirAll(filepath.Join(folder, "consul"), 0755)
	defer func() {
		os.RemoveAll(filepath.Join(folder))
	}()
	mkSecrets(t, folder)
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	consul := &MockConsul{}
	mkObservabilityConfigs(consul, inv, folder, "root")
}
