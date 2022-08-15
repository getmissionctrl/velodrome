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

//go:embed templates/consul/resolved.conf
var resolvedConf string

//go:embed templates/consul/anonymous-policy.hcl
var anonymousDns string

//go:embed templates/consul/install-exporter.sh
var installExporter string

//go:embed templates/consul/consul-exporter.service
var consulExporterService string

//go:embed templates/consul/consul-policies.hcl
var consulPolicies string

//go:embed templates/consul/nomad-client-policy.hcl
var nomadClientPolicy string

//go:embed templates/consul/nomad-server-policy.hcl
var nomadServerPolicy string

//go:embed templates/consul/vault-policy.hcl
var vaultPolicy string

//go:embed templates/consul/prometheus-policy.hcl
var prometheusPolicy string

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

//go:embed templates/vault/vault.service
var vaultService string

//go:embed templates/vault/config.hcl
var vaultConf string

//go:embed templates/ansible/base.yml
var baseAnsible string

//go:embed templates/ansible/consul.yml
var consulAnsible string

//go:embed templates/ansible/nomad.yml
var nomadAnsible string

//go:embed templates/ansible/vault.yml
var vaultAnsible string

//calculate bootstrap expect from files
func Configure(inventoryFile, baseDir, dcName string) error {

	err := os.MkdirAll(filepath.Join(baseDir), 0750)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "base.yml"), []byte(strings.ReplaceAll(baseAnsible, "dc1", dcName)), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "consul.yml"), []byte(strings.ReplaceAll(consulAnsible, "dc1", dcName)), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "nomad.yml"), []byte(strings.ReplaceAll(nomadAnsible, "dc1", dcName)), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "vault.yml"), []byte(strings.ReplaceAll(vaultAnsible, "dc1", dcName)), 0600)
	if err != nil {
		return err
	}

	inv, err := readInventory(inventoryFile)
	if err != nil {
		return err
	}
	err = makeConsulPolicies(inv, baseDir)
	if err != nil {
		return err
	}
	err = makeConfigs(inv, baseDir, dcName)
	if err != nil {
		return err
	}
	err = Secrets(inv, baseDir, dcName)
	return err
}

func readInventory(inventoryFile string) (*aini.InventoryData, error) {

	f, err := os.Open(filepath.Clean(inventoryFile))
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

func getSecrets(baseDir string) (*secretsConfig, error) {
	bytes, err := ioutil.ReadFile(filepath.Clean(filepath.Join(baseDir, "secrets", "secrets.yml")))
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

func writeSecrets(baseDir string, secrets *secretsConfig) error {
	bytes, err := yaml.Marshal(secrets)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "secrets", "secrets.yml"), bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func makeConsulPolicies(inv *aini.InventoryData, baseDir string) error {

	err := os.MkdirAll(filepath.Join(baseDir, "consul"), 0750)
	if err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(baseDir, "consul", "consul-policies.hcl"))

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
	err = os.WriteFile(filepath.Join(baseDir, "consul", "consul-policies.hcl"), output, 0600)
	if err != nil {
		return err
	}

	toCopy := map[string]string{
		filepath.Join(baseDir, "consul", "nomad-client-policy.hcl"): nomadClientPolicy,
		filepath.Join(baseDir, "consul", "install-exporter.sh"):     installExporter,
		filepath.Join(baseDir, "consul", "consul-exporter.service"): consulExporterService,
		filepath.Join(baseDir, "consul", "nomad-server-policy.hcl"): nomadServerPolicy,
		filepath.Join(baseDir, "consul", "prometheus-policy.hcl"):   prometheusPolicy,
		filepath.Join(baseDir, "consul", "anonymous-policy.hcl"):    anonymousDns,
		filepath.Join(baseDir, "consul", "vault-policy.hcl"):        vaultPolicy,
	}
	for k, v := range toCopy {
		err = os.WriteFile(k, []byte(v), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

func getHosts(inv *aini.InventoryData, group string) []string {
	hosts := []string{}
	for k, v := range inv.Groups {
		if k == group {
			for hostname := range v.Hosts {
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

func getPrivateHostNames(inv *aini.InventoryData, group string) []string {
	hosts := []string{}
	for k, v := range inv.Groups {
		if k == group {
			for _, h := range v.Hosts {
				hosts = append(hosts, h.Vars["host_name"])
			}
			return hosts
		}
	}
	return hosts
}

func makeConfigs(inv *aini.InventoryData, baseDir, dcName string) error {
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
	err := os.WriteFile(filepath.Join(baseDir, "consul", "client.j2"), []byte(clientWithDC), 0600)
	if err != nil {
		return err
	}

	serverWithDC := strings.ReplaceAll(consulServer, "dc1", dcName)
	serverWithDC = strings.ReplaceAll(serverWithDC, "join_servers", hosts)
	serverWithDC = strings.ReplaceAll(serverWithDC, "EXPECTS_NO", fmt.Sprintf("%v", len(getHosts(inv, "consul_servers"))))
	err = os.WriteFile(filepath.Join(baseDir, "consul", "server.j2"), []byte(serverWithDC), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "consul", "resolved.conf"), []byte(resolvedConf), 0600)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(baseDir, "nomad"), 0750)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(baseDir, "vault"), 0750)
	if err != nil {
		return err
	}
	nomadServerService := strings.ReplaceAll(nomadService, "nomad_user", "nomad")
	nomadClientService := strings.ReplaceAll(nomadService, "nomad_user", "root")

	nomadServer = strings.ReplaceAll(nomadServer, "EXPECTS_NO", fmt.Sprintf("%v", len(getHosts(inv, "nomad_servers"))))
	nomadServer = strings.ReplaceAll(nomadServer, "dc1", dcName)
	nomadClient = strings.ReplaceAll(nomadClient, "dc1", dcName)

	toWrite := map[string]string{
		filepath.Join(baseDir, "nomad", "server.j2"):            nomadServer,
		filepath.Join(baseDir, "nomad", "client.j2"):            nomadClient,
		filepath.Join(baseDir, "nomad", "nomad-server.service"): nomadServerService,
		filepath.Join(baseDir, "nomad", "nomad-client.service"): nomadClientService,
		filepath.Join(baseDir, "consul", "consul.service"):      consulService,
		filepath.Join(baseDir, "vault", "vault.service"):        vaultService,
		filepath.Join(baseDir, "vault", "config.hcl"):           vaultConf,
	}

	for k, v := range toWrite {
		err = os.WriteFile(k, []byte(v), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func Secrets(inv *aini.InventoryData, baseDir, dcName string) error {
	var out bytes.Buffer
	err := runCmd("", "consul keygen", &out)
	if err != nil {
		return err
	}
	consulSecretDir := filepath.Join(baseDir, "secrets", "consul")
	nomadSecretDir := filepath.Join(baseDir, "secrets", "nomad")
	err = os.MkdirAll(consulSecretDir, 0750)
	if err != nil {
		return err
	}
	consulGossipKey := strings.ReplaceAll(out.String(), "\n", "")

	var out2 bytes.Buffer
	err = runCmd("", "nomad operator keygen", &out2)

	if err != nil {
		return err
	}
	nomadGossipKey := strings.ReplaceAll(out2.String(), "\n", "")
	if os.Getenv("S3_ENDPOINT") == "" || os.Getenv("S3_SECRET_KEY") == "" || os.Getenv("S3_ACCESS_KEY") == "" {
		return fmt.Errorf("s3 compatible env variables missing for storing state: please set S3_ENDPOINT, S3_SECRET_KEY & S3_ACCESS_KEY")
	}

	secrets := &secretsConfig{
		ConsulGossipKey:        consulGossipKey,
		NomadGossipKey:         nomadGossipKey,
		NomadClientConsulToken: "TBD",
		NomadServerConsulToken: "TBD",
		ConsulAgentToken:       "TBD",
		ConsulBootstrapToken:   "TBD",
		S3Endpoint:             os.Getenv("S3_ENDPOINT"),
		S3SecretKey:            os.Getenv("S3_SECRET_KEY"),
		S3AccessKey:            os.Getenv("S3_ACCESS_KEY"),
	}

	if _, err1 := os.Stat(filepath.Join(baseDir, "secrets", "secrets.yml")); errors.Is(err1, os.ErrNotExist) {
		d, e := yaml.Marshal(&secrets)
		if e != nil {
			return e
		}
		e = os.WriteFile(filepath.Join(baseDir, "secrets", "secrets.yml"), d, 0600)
		if e != nil {
			return e
		}
	}

	if err != nil {
		return err
	}
	err = os.MkdirAll(nomadSecretDir, 0750)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(baseDir, "secrets", "consul", "consul-agent-ca.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(consulSecretDir, "consul tls ca create", os.Stdout)
		if err != nil {
			return err
		}
		err = runCmd(consulSecretDir, fmt.Sprintf("consul tls cert create -server -dc %s", dcName), os.Stdout)
		if err != nil {
			return err
		}

	}

	if _, err := os.Stat(filepath.Join(baseDir, "secrets", "nomad", "cli.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(nomadSecretDir, "cfssl print-defaults csr | cfssl gencert -initca - | cfssljson -bare nomad-ca", os.Stdout)
		if err != nil {
			return err
		}
		hosts := getHosts(inv, "nomad_servers")
		privateHosts := getPrivateHosts(inv, "nomad_servers")
		hostString := fmt.Sprintf("server.global.nomad,%s,%s", strings.Join(hosts, ","), strings.Join(privateHosts, ","))
		fmt.Println("generating cert for hosts: " + hostString)

		err = os.WriteFile(filepath.Join(nomadSecretDir, "cfssl.json"), []byte(cfssl), 0600)
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

	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
