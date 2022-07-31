package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMkObservabilityConfigs(t *testing.T) {
	folder := RandString(8)
	assert.NoError(t, os.MkdirAll(filepath.Join(folder, "secrets"), 0750))

	assert.NoError(t, os.MkdirAll(filepath.Join(folder, "consul"), 0750))
	defer func() {
		assert.NoError(t, os.RemoveAll(filepath.Join(folder)))
	}()
	mkSecrets(t, folder)
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	consul := &MockConsul{}
	assert.NoError(t, mkObservabilityConfigs(consul, inv, folder))

	assert.Equal(t, 5, len(consul.RegisterIntentionCalls()))
	assert.Equal(t, 5, len(consul.RegisterServiceCalls()))

	assert.Equal(t, 23, readDir(folder))
}

func readDir(str string) int {
	count := 0
	d, e := os.ReadDir(str)
	if e != nil {
		panic(e)
	}
	for _, f := range d {
		if f.IsDir() {
			count = count + readDir(filepath.Join(str, f.Name()))
		} else {
			count++
		}
	}
	return count
}
