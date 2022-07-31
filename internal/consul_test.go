package internal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestParseConsulToken(t *testing.T) {

	token, err := parseConsulToken(filepath.Join("testdata", "bootstrap.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "4456269a-e46a-c5bd-08d5-914552161f02", token)

}

func TestBootstrapConsul(t *testing.T) {
	folder := RandString(8)
	assert.NoError(t, os.MkdirAll(filepath.Clean(filepath.Join(folder, "secrets")), 0755))

	assert.NoError(t, os.MkdirAll(filepath.Clean(filepath.Join(folder, "consul")), 0755))
	defer func() {
		err := os.RemoveAll(filepath.Join(folder))
		assert.NoError(t, err)
	}()

	mkSecrets(t, folder)

	consul := &MockConsul{
		BootstrapFunc: func() (string, error) {
			return "bootstrap-token", nil
		},
		RegisterACLFunc: func(description, policy string) (string, error) {
			return policy, nil
		},
	}
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	b, err := BootstrapConsul(consul, inv, folder)
	assert.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, 5, len(consul.RegisterPolicyCalls()))

	assert.Equal(t, 4, len(consul.RegisterACLCalls()))
	newSecrets, err := getSecrets(folder)
	assert.NoError(t, err)

	assert.Equal(t, newSecrets.ConsulBootstrapToken, "bootstrap-token")
	assert.Equal(t, newSecrets.ConsulAgentToken, "consul-policies")
	assert.Equal(t, newSecrets.NomadClientConsulToken, "nomad-client")
	assert.Equal(t, newSecrets.NomadServerConsulToken, "nomad-server")
}

func mkSecrets(t *testing.T, folder string) *secretsConfig {
	secrets := &secretsConfig{
		ConsulGossipKey:        "consulGossipKey",
		NomadGossipKey:         "nomadGossipKey",
		NomadClientConsulToken: "TBD",
		NomadServerConsulToken: "TBD",
		ConsulAgentToken:       "TBD",
		ConsulBootstrapToken:   "TBD",
		S3Endpoint:             "s3_endpoint_test",
		S3SecretKey:            "s3_secret_key_test",
		S3AccessKey:            "s3_access_key_test",
	}

	if _, err := os.Stat(filepath.Join(folder, "secrets", "secrets.yml")); errors.Is(err, os.ErrNotExist) {
		d, err := yaml.Marshal(&secrets)
		assert.NoError(t, err)
		err = os.WriteFile(filepath.Join(folder, "secrets", "secrets.yml"), d, 0755)
		assert.NoError(t, err)
	}
	return secrets
}
