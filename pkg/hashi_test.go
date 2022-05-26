package pkg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMakeConsulPolicies(t *testing.T) {
	inv, err := readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	err = makeConsulPolicies(inv)
	assert.NoError(t, err)

	defer func() {
		os.RemoveAll(filepath.Join("config", "consul"))
	}()

	bytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "consul-policies.hcl"))
	assert.NoError(t, err)

	contents := string(bytes)

	assert.Contains(t, contents, `node "ubuntu-4gb-nbg1-1"`)

	assert.Equal(t, 5, strings.Count(contents, "node"))

	_, err = ioutil.ReadFile(filepath.Join("config", "consul", "nomad-server-policy.hcl"))
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join("config", "consul", "nomad-client-policy.hcl"))
	assert.NoError(t, err)

	inv, err = readInventory(filepath.Join("testdata", "inventory"))
	assert.NoError(t, err)
	makeConsulConfig("hetzner", inv)

	serverBytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "server.j2"))
	assert.NoError(t, err)
	serverConf := string(serverBytes)

	clientBytes, err := ioutil.ReadFile(filepath.Join("config", "consul", "client.j2"))
	assert.NoError(t, err)
	clientConf := string(clientBytes)

	assert.Contains(t, clientConf, `datacenter = "hetzner"`)
	assert.Contains(t, serverConf, `datacenter = "hetzner"`)
	retryJoin := `"10.0.0.3"`
	assert.Contains(t, clientConf, retryJoin)
	assert.Contains(t, serverConf, retryJoin)
	assert.Contains(t, serverConf, `key_file = "/etc/consul.d/certs/hetzner-server-consul-0-key.pem"`)

}

func TestMakeSecrets(t *testing.T) {
	// defer func() {
	// 	os.RemoveAll(filepath.Join("config", "secrets"))
	// }()
	err := Secrets("dc1")
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(filepath.Join("config", "secrets", "secrets.yml"))
	assert.NoError(t, err)
	err = Secrets("d1c1")
	assert.NoError(t, err)
	bytes2, err := ioutil.ReadFile(filepath.Join("config", "secrets", "secrets.yml"))
	assert.NoError(t, err)
	conf := string(bytes)

	assert.Equal(t, conf, string(bytes2))

	assert.Equal(t, 1, strings.Count(conf, "CONSUL_GOSSIP_KEY"))
	assert.Equal(t, 1, strings.Count(conf, "NOMAD_GOSSIP_KEY"))

	var theMap map[string]interface{}
	err = yaml.Unmarshal([]byte(bytes), &theMap)
	assert.NoError(t, err)
	_, ok := theMap["CONSUL_GOSSIP_KEY"]
	assert.True(t, ok)
	_, ok = theMap["NOMAD_GOSSIP_KEY"]
	assert.True(t, ok)

	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "consul-agent-ca-key.pem"))
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "consul-agent-ca.pem"))
	assert.NoError(t, err)
	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "dc1-server-consul-0-key.pem"))
	assert.NoError(t, err)
	_, err = ioutil.ReadFile(filepath.Join("config", "secrets", "consul", "dc1-server-consul-0.pem"))
	assert.NoError(t, err)
}
