package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMakeConsulPoliciesAndHashiConfigs(t *testing.T) {
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	err = makeConsulPolicies(inv)
	assert.NoError(t, err)

	defer func() {
		os.RemoveAll(filepath.Join("config", "consul"))
		os.RemoveAll(filepath.Join("config", "nomad"))
	}()

	bytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "consul-policies.hcl"))
	assert.NoError(t, err)

	contents := string(bytes)

	assert.Contains(t, contents, `node "ubuntu1"`)

	assert.Equal(t, 5, strings.Count(contents, "node"))

	_, err = ioutil.ReadFile(filepath.Join("config", "consul", "nomad-server-policy.hcl"))
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join("config", "consul", "nomad-client-policy.hcl"))
	assert.NoError(t, err)

	inv, err = readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	makeConfigs(inv, "hetzner", false)

	serverBytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "server.j2"))
	assert.NoError(t, err)
	serverConf := string(serverBytes)

	clientBytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "client.j2"))
	assert.NoError(t, err)
	clientConf := string(clientBytes)

	assertFileExists(t, filepath.Join("config", "nomad", "client.j2"))
	assertFileExists(t, filepath.Join("config", "nomad", "server.j2"))
	assertFileExists(t, filepath.Join("config", "nomad", "nomad-server.service"))

	assertFileExists(t, filepath.Join("config", "nomad", "nomad-client.service"))

	assert.Contains(t, clientConf, `datacenter = "hetzner"`)
	assert.Contains(t, serverConf, `datacenter = "hetzner"`)
	retryJoin := `"10.0.0.3"`
	assert.Contains(t, clientConf, retryJoin)
	assert.Contains(t, serverConf, retryJoin)
	assert.Contains(t, serverConf, `key_file = "/etc/consul.d/certs/hetzner-server-consul-0-key.pem"`)

}

func TestMakeSecrets(t *testing.T) {
	defer func() {
		os.RemoveAll(filepath.Join("config", "secrets"))
	}()
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	err = Secrets(inv, "dc1")
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(filepath.Join("config", "secrets", "secrets.yml"))
	assert.NoError(t, err)
	err = Secrets(inv, "dc1")
	assert.NoError(t, err)
	bytes2, err := ioutil.ReadFile(filepath.Join("config", "secrets", "secrets.yml"))
	assert.NoError(t, err)
	conf := string(bytes)

	fmt.Println(conf)
	fmt.Println("******")
	fmt.Println(string(bytes2))

	assert.Equal(t, conf, string(bytes2))

	assert.Equal(t, 1, strings.Count(conf, "CONSUL_GOSSIP_KEY"))
	assert.Equal(t, 1, strings.Count(conf, "NOMAD_GOSSIP_KEY"))

	var theMap map[string]string
	err = yaml.Unmarshal([]byte(bytes), &theMap)
	assert.NoError(t, err)
	assert.NotEmpty(t, theMap["CONSUL_GOSSIP_KEY"])

	assert.NotEmpty(t, theMap["NOMAD_GOSSIP_KEY"])

	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "consul-agent-ca-key.pem"))
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "consul-agent-ca.pem"))
	assert.NoError(t, err)
	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "dc1-server-consul-0-key.pem"))
	assert.NoError(t, err)
	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "dc1-server-consul-0.pem"))
	assert.NoError(t, err)

	nomadDir := filepath.Join("config", "secrets", "nomad")
	files := []string{
		"cfssl.json",
		"nomad-ca.csr",
		"nomad-ca-key.pem",
		"nomad-ca.pem",
		"cli.csr",
		"cli-key.pem",
		"cli.pem",
		"client.csr",
		"client-key.pem",
		"client.pem",
		"server.csr",
		"server-key.pem",
		"server.pem",
	}

	for _, path := range files {
		assertFileExists(t, filepath.Join(nomadDir, path))
	}

	secrets, err := getSecrets()
	assert.NoError(t, err)
	assert.Equal(t, "TBD", secrets.ConsulBootstrapToken)
	assert.Equal(t, "TBD", secrets.ConsulAgentToken)
	assert.Equal(t, "TBD", secrets.NomadClientConsulToken)
	assert.Equal(t, "TBD", secrets.NomadServerConsulToken)
}

func assertFileExists(t *testing.T, path string) {
	_, err := ioutil.ReadFile(path)
	assert.NoError(t, err, path)
}
