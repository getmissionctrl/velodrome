package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relex/aini"
)

type Consul interface {
	Bootstrap() (string, error)
	RegisterACL(description, policy string) (string, error)
	UpdateACL(tokenID, policy string) error
	UpdatePolicy(name, file string) error
	RegisterPolicy(name, file string) error
	RegisterIntention(file string) error
	RegisterService(file string) error
}

type consulBinary struct {
	inventory *aini.InventoryData
	secrets   *secretsConfig
	baseDir   string
}

func NewConsul(inventory *aini.InventoryData, secrets *secretsConfig, baseDir string) Consul {
	return &consulBinary{inventory: inventory, secrets: secrets, baseDir: baseDir}
}

func (client *consulBinary) Bootstrap() (string, error) {
	hosts := getHosts(client.inventory, "consul_servers")
	if len(hosts) == 0 {
		return "", fmt.Errorf("no consul servers found in inventory")
	}
	host := hosts[0]
	secretsDir := filepath.Join(client.baseDir, "secrets")

	path := filepath.Join(secretsDir, "consul-bootstrap.token")
	exports := fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_CLIENT_CERT=%s/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=%s/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host, client.baseDir, client.baseDir)

	err := runCmd("", fmt.Sprintf(`%s consul acl bootstrap > %s`, exports, path), os.Stdout)
	if err != nil {
		return "", err
	}
	token, err := parseConsulToken(path)
	if err != nil {
		return "", err
	}
	client.secrets.ConsulBootstrapToken = token
	return token, nil
}

func (client *consulBinary) RegisterACL(description, policy string) (string, error) {
	exports, err := client.getExports()
	if err != nil {
		return "", err
	}
	tokenPath := filepath.Join(client.baseDir, "secrets", fmt.Sprintf("%s.token", policy))
	err = runCmd("", fmt.Sprintf(`%sconsul acl token create -description "%s"  -policy-name %s > %s`, exports, description, policy, tokenPath), os.Stdout)
	if err != nil {
		return "", err
	}
	return parseConsulToken(tokenPath)
}

func (client *consulBinary) UpdateACL(tokenID, policy string) error {
	exports, err := client.getExports()
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul acl token update -id %s -policy-name=%s`, exports, tokenID, policy), os.Stdout)
}

func (client *consulBinary) RegisterPolicy(name, file string) error {
	exports, err := client.getExports()
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul acl policy create -name %s -rules @%s`, exports, name, file), os.Stdout)
}

func (client *consulBinary) UpdatePolicy(name, file string) error {
	exports, err := client.getExports()
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul acl policy update -name %s -rules @%s`, exports, name, file), os.Stdout)
}

func (client *consulBinary) RegisterIntention(file string) error {
	exports, err := client.getExports()
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul config write %s`, exports, file), os.Stdout)
}

func (client *consulBinary) RegisterService(file string) error {
	exports, err := client.getExports()
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul services register %s`, exports, file), os.Stdout)
}

func (client *consulBinary) getExports() (string, error) {
	hosts := getHosts(client.inventory, "consul_servers")
	if len(hosts) == 0 {
		return "", fmt.Errorf("no consul servers found in inventory")
	}
	host := hosts[0]

	token := client.secrets.ConsulBootstrapToken
	exports := fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_HTTP_TOKEN="%s" && export CONSUL_CLIENT_CERT=%s/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=%s/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host, token, client.baseDir, client.baseDir)
	return exports, nil
}
