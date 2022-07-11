package internal

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/relex/aini"
	"gopkg.in/yaml.v3"
)

func parseConsulToken(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := string(content)
	temp := strings.Split(text, "\n")
	for _, line := range temp {
		if strings.HasPrefix(line, "SecretID:") {
			return strings.ReplaceAll(strings.ReplaceAll(line, "SecretID:", ""), " ", ""), nil
		}
	}
	return "", nil
}

func registerConsul(inventory *aini.InventoryData, secrets *secretsConfig, file string) error {
	exports, err := getExports(inventory, secrets)
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul services register %s`, exports, file), os.Stdout)
}

func registerIntention(inventory *aini.InventoryData, secrets *secretsConfig, file string) error {
	exports, err := getExports(inventory, secrets)
	if err != nil {
		return err
	}
	return runCmd("", fmt.Sprintf(`%sconsul config write %s`, exports, file), os.Stdout)
}

func getExports(inventory *aini.InventoryData, secrets *secretsConfig) (string, error) {
	hosts := getHosts(inventory, "consul_servers")
	if len(hosts) == 0 {
		return "", fmt.Errorf("no consul servers found in inventory")
	}
	host := hosts[0]
	token := secrets.ConsulBootstrapToken
	exports := fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_HTTP_TOKEN="%s" && export CONSUL_CLIENT_CERT=config/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=config/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host, token)
	return exports, nil
}

func regenerateConsulPolicies(inventory *aini.InventoryData, secrets *secretsConfig) error {
	token := secrets.ConsulBootstrapToken
	hosts := getHosts(inventory, "consul_servers")
	if len(hosts) == 0 {
		return fmt.Errorf("no consul servers found in inventory")
	}
	host := hosts[0]

	err := makeConsulPolicies(inventory)
	if err != nil {
		return err
	}
	fmt.Println("Updating consul policies")
	exports := fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_HTTP_TOKEN="%s" && export CONSUL_CLIENT_CERT=config/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=config/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host, token)
	policyConsul := filepath.Join("config", "consul", "consul-policies.hcl")
	err = runCmd("", fmt.Sprintf(`%sconsul acl policy update -name consul-policies -rules @%s`, exports, policyConsul), os.Stdout)

	return err
}

func BootstrapConsul(inventory string) (bool, error) {
	secrets, err := getSecrets()
	if err != nil {
		return false, err
	}
	inv, err := readInventory(inventory)
	if err != nil {
		return false, err
	}
	if secrets.ConsulBootstrapToken != "TBD" {
		err = regenerateConsulPolicies(inv, secrets)
		return false, err
	}
	hosts := getHosts(inv, "consul_servers")
	if len(hosts) == 0 {
		return false, fmt.Errorf("no consul servers found in inventory")
	}
	host := hosts[0]
	secretsDir := filepath.Join("config", "secrets")

	path := filepath.Join(secretsDir, "consul-bootstrap.token")
	exports := fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_CLIENT_CERT=config/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=config/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host)

	err = runCmd("", fmt.Sprintf(`%s consul acl bootstrap > %s`, exports, path), os.Stdout)
	if err != nil {
		return false, err
	}
	token, err := parseConsulToken(path)
	if err != nil {
		return false, err
	}
	secrets.ConsulBootstrapToken = token
	exports = fmt.Sprintf(`export CONSUL_HTTP_ADDR="%s:8501" && export CONSUL_HTTP_TOKEN="%s" && export CONSUL_CLIENT_CERT=config/secrets/consul/consul-agent-ca.pem && export CONSUL_CLIENT_KEY=config/secrets/consul/consul-agent-ca-key.pem && export CONSUL_HTTP_SSL=true && export CONSUL_HTTP_SSL_VERIFY=false && `, host, token)
	policyConsul := filepath.Join("config", "consul", "consul-policies.hcl")
	err = runCmd("", fmt.Sprintf(`%sconsul acl policy create -name consul-policies -rules @%s`, exports, policyConsul), os.Stdout)
	if err != nil {
		return false, err
	}
	policyConsul = filepath.Join("config", "consul", "nomad-client-policy.hcl")
	err = runCmd("", fmt.Sprintf(`%sconsul acl policy create -name nomad-client -rules @%s`, exports, policyConsul), os.Stdout)
	if err != nil {
		return false, err
	}
	policyConsul = filepath.Join("config", "consul", "nomad-server-policy.hcl")
	err = runCmd("", fmt.Sprintf(`%sconsul acl policy create -name nomad-server -rules @%s`, exports, policyConsul), os.Stdout)
	if err != nil {
		return false, err
	}

	policyConsul = filepath.Join("config", "consul", "anonymous-policy.hcl")
	err = runCmd("", fmt.Sprintf(`%sconsul acl policy create -name anonymous-dns-read -rules @%s`, exports, policyConsul), os.Stdout)
	if err != nil {
		return false, err
	}

	err = runCmd("", fmt.Sprintf(`%sconsul acl token update -id anonymous -policy-name=anonymous-dns-read`, exports), os.Stdout)
	if err != nil {
		return false, err
	}

	tokenPath := filepath.Join(secretsDir, "consul-client.token")
	err = runCmd("", fmt.Sprintf(`%sconsul acl token create -description "agent token"  -policy-name consul-policies > %s`, exports, tokenPath), os.Stdout)
	if err != nil {
		return false, err
	}
	clientToken, err := parseConsulToken(tokenPath)
	if err != nil {
		return false, err
	}

	secrets.ConsulAgentToken = clientToken

	tokenPath = filepath.Join(secretsDir, "nomad-client.token")
	err = runCmd("", fmt.Sprintf(`%sconsul acl token create -description "nomad client token"  -policy-name nomad-client > %s`, exports, tokenPath), os.Stdout)
	if err != nil {
		return false, err
	}
	clientToken, err = parseConsulToken(tokenPath)
	if err != nil {
		return false, err
	}

	secrets.NomadClientConsulToken = clientToken

	tokenPath = filepath.Join(secretsDir, "nomad-server.token")
	err = runCmd("", fmt.Sprintf(`%sconsul acl token create -description "nomad server token"  -policy-name nomad-server > %s`, exports, tokenPath), os.Stdout)
	if err != nil {
		return false, err
	}
	clientToken, err = parseConsulToken(tokenPath)
	if err != nil {
		return false, err
	}

	secrets.NomadServerConsulToken = clientToken
	d, err := yaml.Marshal(&secrets)
	if err != nil {
		return false, err
	}
	err = os.WriteFile(filepath.Join("config", "secrets", "secrets.yml"), d, 0755)

	return true, err
}
