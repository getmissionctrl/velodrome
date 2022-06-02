package internal

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/relex/aini"
	"gopkg.in/yaml.v3"
)

//go:embed templates/consul/consul-policies.hcl
var consulPolicies string

//go:embed templates/consul/nomad-client-policy.hcl
var nomadClientPolicy string

//go:embed templates/consul/nomad-server-policy.hcl
var nomadServerPolicy string

//go:embed templates/consul/consul-server-config.hcl
var consulServer string

//go:embed templates/consul/consul-client-config.hcl
var consulClient string

//go:embed templates/nomad/cfssl.json
var cfssl string

//go:embed templates/nomad/client.j2
var nomadClient string

//go:embed templates/nomad/server.j2
var nomadServer string

//go:embed templates/nomad/nomad.service
var nomadService string

//go:embed templates/consul/consul.service
var consulService string

//go:embed templates/ansible/setup.yml
var setupAnsible string

//go:embed templates/ansible/destroy.yml
var destroyAnsible string

type secretsConfig struct {
	ConsulGossipKey        string `yaml:"CONSUL_GOSSIP_KEY"`
	NomadGossipKey         string `yaml:"NOMAD_GOSSIP_KEY"`
	NomadClientConsulToken string `yaml:"NOMAD_CLIENT_CONSUL_TOKEN"`
	NomadServerConsulToken string `yaml:"NOMAD_SERVER_CONSUL_TOKEN"`
	ConsulAgentToken       string `yaml:"CONSUL_AGENT_TOKEN"`
	ConsulBootstrapToken   string `yaml:"CONSUL_BOOTSTRAP_TOKEN"`
}

//calculate bootstrap expect from files
func Configure(inventoryFile, dcName string) error {

	err := os.MkdirAll(filepath.Join("config"), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "setup.yml"), []byte(strings.ReplaceAll(setupAnsible, "dc1", dcName)), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "destroy.yml"), []byte(destroyAnsible), 0755)
	if err != nil {
		return err
	}

	inv, err := readInventory(inventoryFile)
	if err != nil {
		return err
	}
	err = makeConsulPolicies(inv)
	if err != nil {
		return err
	}
	err = makeConfigs(inv, dcName)
	if err != nil {
		return err
	}
	err = Secrets(inv, dcName)
	return err
}

func readInventory(inventoryFile string) (*aini.InventoryData, error) {

	f, err := os.Open(inventoryFile)
	defer func() {
		e := f.Close()
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}
	}()

	if err != nil {
		return nil, err
	}

	return aini.Parse(f)
}

func getSecrets() (*secretsConfig, error) {
	bytes, err := ioutil.ReadFile(filepath.Join("config", "secrets", "secrets.yml"))
	if err != nil {
		return nil, err
	}
	var secrets secretsConfig
	err = yaml.Unmarshal(bytes, &secrets)
	if err != nil {
		return nil, err
	}
	return &secrets, nil

}

func makeConsulPolicies(inv *aini.InventoryData) error {

	err := os.MkdirAll(filepath.Join("config", "consul"), 0755)
	if err != nil {
		return err
	}
	_ = os.Remove(filepath.Join("config", "consul", "consul-policies.hcl"))

	hostMap := make(map[string]string)
	hosts := []string{}
	for _, v := range inv.Groups {
		for _, v := range v.Hosts {
			if _, ok := hostMap[v.Vars["host_name"]]; !ok {
				hosts = append(hosts, v.Vars["host_name"])
				hostMap[v.Vars["host_name"]] = v.Vars["host_name"]
			}
		}
	}

	tmpl, e := template.New("consul-policies").Parse(consulPolicies)
	if e != nil {
		return e
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string][]string{"Hosts": hosts})
	if err != nil {
		return err
	}

	output := buf.Bytes()
	err = os.WriteFile(filepath.Join("config", "consul", "consul-policies.hcl"), output, 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("config", "consul", "nomad-client-policy.hcl"), []byte(nomadClientPolicy), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("config", "consul", "nomad-server-policy.hcl"), []byte(nomadServerPolicy), 0755)
	if err != nil {
		return err
	}

	return nil
}

func getHosts(inv *aini.InventoryData, group string) []string {
	hosts := []string{}
	for k, v := range inv.Groups {
		if k == group {
			for hostname, _ := range v.Hosts {
				hosts = append(hosts, hostname)
			}
			return hosts
		}
	}
	return hosts
}

func getPrivateHosts(inv *aini.InventoryData, group string) []string {
	hosts := []string{}
	for k, v := range inv.Groups {
		if k == group {
			for _, h := range v.Hosts {
				hosts = append(hosts, h.Vars["private_ip"])
			}
			return hosts
		}
	}
	return hosts
}

func makeConfigs(inv *aini.InventoryData, dcName string) error {
	hostMap := make(map[string]string)
	hosts := ""
	first := true
	for k, v := range inv.Groups {
		if k == "consul_servers" {
			for _, v := range v.Hosts {
				if _, ok := hostMap[v.Vars["private_ip"]]; !ok {
					if first {
						hosts = fmt.Sprintf(`"%v"`, v.Vars["private_ip"])
						first = false
					} else {
						hosts = hosts + `, ` + fmt.Sprintf(`"%v"`, v.Vars["private_ip"])
					}
					hostMap[v.Vars["private_ip"]] = v.Vars["private_ip"]
				}
			}
		}
	}
	clientWithDC := strings.ReplaceAll(consulClient, "dc1", dcName)
	clientWithDC = strings.ReplaceAll(clientWithDC, "join_servers", hosts)
	err := os.WriteFile(filepath.Join("config", "consul", "client.j2"), []byte(clientWithDC), 0755)
	if err != nil {
		return err
	}

	serverWithDC := strings.ReplaceAll(consulServer, "dc1", dcName)
	serverWithDC = strings.ReplaceAll(serverWithDC, "join_servers", hosts)
	serverWithDC = strings.ReplaceAll(serverWithDC, "EXPECTS_NO", fmt.Sprintf("%v", len(getHosts(inv, "consul_servers"))))
	err = os.WriteFile(filepath.Join("config", "consul", "server.j2"), []byte(serverWithDC), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join("config", "nomad"), 0755)
	if err != nil {
		return err
	}
	nomadServerService := strings.ReplaceAll(nomadService, "nomad_user", "nomad")
	nomadClientService := strings.ReplaceAll(nomadService, "nomad_user", "root")

	nomadServer = strings.ReplaceAll(nomadServer, "EXPECTS_NO", fmt.Sprintf("%v", len(getHosts(inv, "nomad_servers"))))

	err = os.WriteFile(filepath.Join("config", "nomad", "server.j2"), []byte(nomadServer), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("config", "nomad", "client.j2"), []byte(nomadClient), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("config", "nomad", "nomad-server.service"), []byte(nomadServerService), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join("config", "nomad", "nomad-client.service"), []byte(nomadClientService), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "consul", "consul.service"), []byte(consulService), 0755)
	if err != nil {
		return err
	}

	return nil
}

func Secrets(inv *aini.InventoryData, dcName string) error {
	var out bytes.Buffer
	err := runCmd("", "consul keygen", &out)
	if err != nil {
		return err
	}
	consulSecretDir := filepath.Join("config", "secrets", "consul")
	nomadSecretDir := filepath.Join("config", "secrets", "nomad")
	err = os.MkdirAll(consulSecretDir, 0755)
	consulGossipKey := strings.ReplaceAll(string(out.Bytes()), "\n", "")

	var out2 bytes.Buffer
	err = runCmd("", "nomad operator keygen", &out2)

	if err != nil {
		return err
	}
	nomadGossipKey := strings.ReplaceAll(string(out2.Bytes()), "\n", "")

	secrets := &secretsConfig{
		ConsulGossipKey:        consulGossipKey,
		NomadGossipKey:         nomadGossipKey,
		NomadClientConsulToken: "TBD",
		NomadServerConsulToken: "TBD",
		ConsulAgentToken:       "TBD",
		ConsulBootstrapToken:   "TBD",
	}

	if _, err := os.Stat(filepath.Join("config", "secrets", "secrets.yml")); errors.Is(err, os.ErrNotExist) {
		d, err := yaml.Marshal(&secrets)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join("config", "secrets", "secrets.yml"), d, 0755)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	err = os.MkdirAll(nomadSecretDir, 0755)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join("config", "secrets", "consul", "consul-agent-ca.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(consulSecretDir, "consul tls ca create", os.Stdout)
		if err != nil {
			return err
		}
		err = runCmd(consulSecretDir, fmt.Sprintf("consul tls cert create -server -dc %s", dcName), os.Stdout)
		if err != nil {
			return err
		}

	}

	if _, err := os.Stat(filepath.Join("config", "secrets", "nomad", "cli.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(nomadSecretDir, "cfssl print-defaults csr | cfssl gencert -initca - | cfssljson -bare nomad-ca", os.Stdout)
		if err != nil {
			return err
		}
		hosts := getHosts(inv, "nomad_servers")
		privateHosts := getPrivateHosts(inv, "nomad_servers")
		hostString := fmt.Sprintf("server.global.nomad,%s,%s", strings.Join(hosts, ","), strings.Join(privateHosts, ","))
		fmt.Println("generating cert for hosts: " + hostString)

		err = os.WriteFile(filepath.Join(nomadSecretDir, "cfssl.json"), []byte(cfssl), 0755)
		if err != nil {
			return err
		}
		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare server`, hostString), os.Stdout)
		if err != nil {
			return err
		}

		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare client`, hostString), os.Stdout)
		if err != nil {
			return err
		}

		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare cli`, hostString), os.Stdout)
		if err != nil {
			return err
		}

	}
	return nil
}

func runCmd(dir, command string, stdOut io.Writer) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

	cmd.Start()
	return cmd.Wait()
}
