package internal

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/terraform/hetzner/main.tf
var hetznerMain string

//go:embed templates/terraform/hetzner/vars.tf
var hetznerVars string

func GenerateTerraform(config *Config) error {
	settings := map[string]struct {
		Main string
		Vars string
	}{
		"hetzner": {
			Main: hetznerMain,
			Vars: hetznerVars,
		},
	}

	tfSettings, ok := settings[config.CloudProviderConfig.Provider]
	if !ok {
		return fmt.Errorf("%s is not a supported cloud provider", config.CloudProviderConfig.Provider)
	}

	tmpl, e := template.New("tf-vars").Parse(tfSettings.Vars)
	if e != nil {
		return e
	}
	var buf bytes.Buffer

	err := tmpl.Execute(&buf, config)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Clean(filepath.Join(config.BaseDir, "terraform")), 0750)
	if err != nil {
		return err
	}
	folder := filepath.Join(config.BaseDir, "terraform")

	err = os.WriteFile(filepath.Clean(filepath.Join(folder, "vars.tf")), buf.Bytes(), 0600)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Clean(filepath.Join(folder, "main.tf")), []byte(hetznerMain), 0600)
	if err != nil {
		return err
	}
	return nil
}

func GenerateEnvFile() error {
	return nil
}

func GenerateInventory(config *Config) error {
	jsonFile, err := os.Open(filepath.Clean(filepath.Join(config.BaseDir, "inventory-output.json")))
	if err != nil {
		return err
	}
	defer func() {
		e := jsonFile.Close()
		fmt.Println(e)
	}()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	var inventory InventoryJson

	err = json.Unmarshal(byteValue, &inventory)
	if err != nil {
		return err
	}

	if len(inventory.ConsulServers.Value) == 0 {
		inventory.ConsulServers.Value = inventory.NomadServers.Value
	}

	takeFirst := len(inventory.ObservabilityServers.Value) == 1
	if takeFirst {
		inventory.grafanaServers = []string{inventory.ObservabilityServers.Value[0]}
		inventory.prometheusServers = []string{inventory.ObservabilityServers.Value[0]}
		inventory.lokiServers = []string{inventory.ObservabilityServers.Value[0]}
		inventory.tempoServers = []string{inventory.ObservabilityServers.Value[0]}
	} else {
		inventory.grafanaServers = []string{inventory.ObservabilityServers.Value[0]}
		inventory.prometheusServers = []string{inventory.ObservabilityServers.Value[1]}
		inventory.lokiServers = []string{inventory.ObservabilityServers.Value[2]}
		inventory.tempoServers = []string{inventory.ObservabilityServers.Value[3]}
	}

	sections := []struct {
		name   string
		values []string
	}{
		{name: "consul_servers",
			values: inventory.ConsulServers.Value,
		},
		{name: "vault_servers",
			values: inventory.VaultServers.Value,
		},
		{name: "nomad_servers",
			values: inventory.NomadServers.Value,
		},
		{name: "clients",
			values: inventory.Clients.Value,
		},
		{name: "grafana",
			values: inventory.grafanaServers,
		},
		{name: "prometheus",
			values: inventory.prometheusServers,
		},
		{name: "loki",
			values: inventory.lokiServers,
		},
		{name: "tempo",
			values: inventory.tempoServers,
		},
	}

	inventoryStr := ""
	for _, section := range sections {
		if len(section.values) > 0 {
			inventoryStr = fmt.Sprintf("%s[%s]\n", inventoryStr, section.name)
			for _, v := range section.values {
				inventoryStr = fmt.Sprintf("%s%s\n", inventoryStr, v)
			}
		}
	}

	return os.WriteFile(filepath.Clean(filepath.Join(config.BaseDir, "inventory")), []byte(inventoryStr), 0600)

}

type InventoryJson struct {
	Clients              InvValue `json:"client_servers"`
	NomadServers         InvValue `json:"nomad_servers"`
	ObservabilityServers InvValue `json:"o11y_servers"`
	VaultServers         InvValue `json:"vault_servers"`
	ConsulServers        InvValue `json:"consul_servers"`
	prometheusServers    []string
	grafanaServers       []string
	lokiServers          []string
	tempoServers         []string
}
type InvValue struct {
	Value []string `json:"value"`
}
