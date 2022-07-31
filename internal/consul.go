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
	content, err := ioutil.ReadFile(filepath.Clean(file))
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

func regenerateConsulPolicies(consul Consul, inventory *aini.InventoryData, baseDir string) error {
	err := makeConsulPolicies(inventory, baseDir)
	if err != nil {
		return err
	}
	fmt.Println("Updating consul policies")
	policyConsul := filepath.Join(baseDir, "consul", "consul-policies.hcl")

	return consul.UpdatePolicy("consul-policies", policyConsul)
}

func BootstrapConsul(consul Consul, inventory *aini.InventoryData, baseDir string) (bool, error) {
	secrets, err := getSecrets(baseDir)
	if err != nil {
		return false, err
	}

	if secrets.ConsulBootstrapToken != "TBD" {
		err = regenerateConsulPolicies(consul, inventory, baseDir)
		return false, err
	}
	token, err := consul.Bootstrap()
	if secrets.ConsulBootstrapToken != "TBD" {
		return false, err
	}

	secrets.ConsulBootstrapToken = token

	policies := map[string]string{
		"consul-policies":    filepath.Join(baseDir, "consul", "consul-policies.hcl"),
		"nomad-client":       filepath.Join(baseDir, "consul", "nomad-client-policy.hcl"),
		"nomad-server":       filepath.Join(baseDir, "consul", "nomad-server-policy.hcl"),
		"prometheus":         filepath.Join(baseDir, "consul", "prometheus-policy.hcl"),
		"anonymous-dns-read": filepath.Join(baseDir, "consul", "anonymous-policy.hcl"),
	}

	for k, v := range policies {
		err = consul.RegisterPolicy(k, v)
		if err != nil {
			return false, err
		}
	}

	err = consul.UpdateACL("anonymous", "anonymous-dns-read")
	if err != nil {
		return false, err
	}

	acls := map[string]string{
		"agent token":        "consul-policies",
		"client token":       "nomad-client",
		"nomad server token": "nomad-server",
		"prometheus token":   "prometheus",
	}
	tokens := map[string]string{}

	for k, v := range acls {
		clientToken, e := consul.RegisterACL(k, v)
		if e != nil {
			return false, e
		}
		tokens[v] = clientToken
	}

	secrets.ConsulAgentToken = tokens["consul-policies"]
	secrets.NomadClientConsulToken = tokens["nomad-client"]
	secrets.NomadServerConsulToken = tokens["nomad-server"]
	secrets.PrometheusConsulToken = tokens["prometheus"]

	d, err := yaml.Marshal(&secrets)
	if err != nil {
		return false, err
	}
	err = os.WriteFile(filepath.Join(baseDir, "secrets", "secrets.yml"), d, 0755)

	return true, err
}
