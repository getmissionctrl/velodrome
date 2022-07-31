package internal

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

type tfConfig struct {
	Variable []Variable `hcl:"variable,block"`
}

type Variable struct {
	Type      *hcl.Attribute `hcl:"type"`
	Name      string         `hcl:"name,label"`
	Default   *cty.Value     `hcl:"default,optional"`
	Sensitive bool           `hcl:"sensitive,optional"`
}

var inventoryResultTest = `[consul_servers]
116.203.48.99 host_name=nomad-srv-nomad-server-1 private_ip=10.0.1.2
195.201.124.132 host_name=nomad-srv-nomad-server-2 private_ip=10.0.1.3
116.203.21.77 host_name=nomad-srv-nomad-server-3 private_ip=10.0.1.4
[nomad_servers]
116.203.48.99 host_name=nomad-srv-nomad-server-1 private_ip=10.0.1.2
195.201.124.132 host_name=nomad-srv-nomad-server-2 private_ip=10.0.1.3
116.203.21.77 host_name=nomad-srv-nomad-server-3 private_ip=10.0.1.4
[clients]
159.69.147.86 host_name=nomad-srv-client-1 private_ip=10.0.3.2
195.201.127.116 host_name=nomad-srv-client-2 private_ip=10.0.3.3
[grafana]
195.201.126.249 host_name=nomad-srv-observability-1 private_ip=10.0.4.2
[prometheus]
195.201.126.249 host_name=nomad-srv-observability-1 private_ip=10.0.4.2
[loki]
195.201.126.249 host_name=nomad-srv-observability-1 private_ip=10.0.4.2
[tempo]
195.201.126.249 host_name=nomad-srv-observability-1 private_ip=10.0.4.2
`

func TestGenerateTerraform(t *testing.T) {
	config, err := LoadConfig("testdata/config.yaml")
	assert.NoError(t, err)

	folder := RandString(8)
	config.BaseDir = folder
	defer func() {
		err := os.RemoveAll(filepath.Join(folder))
		assert.NoError(t, err)
	}()

	err = GenerateTerraform(config)
	assert.NoError(t, err)

	parser := hclparse.NewParser()
	f, parseDiags := parser.ParseHCLFile(filepath.Join(folder, "terraform", "vars.tf"))
	assert.False(t, parseDiags.HasErrors())

	_, parseDiags = parser.ParseHCLFile(filepath.Join(folder, "terraform", "main.tf"))
	assert.False(t, parseDiags.HasErrors())

	var conf tfConfig
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &conf)
	assert.False(t, decodeDiags.HasErrors())

	vars := []struct {
		name       string
		tpe        string
		defaultVal cty.Value
	}{
		{name: "hcloud_token", tpe: "string", defaultVal: cty.NullVal(cty.String)},
		{name: "server_count", tpe: "number", defaultVal: cty.NumberVal(big.NewFloat(3))},
		{name: "client_count", tpe: "number", defaultVal: cty.NumberVal(big.NewFloat(2))},
		{name: "separate_consul_servers", tpe: "bool", defaultVal: cty.BoolVal(false)},
		{name: "multi_instance_observability", tpe: "bool", defaultVal: cty.BoolVal(false)},
		{name: "ssh_keys", tpe: "list", defaultVal: cty.TupleVal([]cty.Value{cty.StringVal("wille.faler@gmail.com")})},
		{name: "base_server_name", tpe: "string", defaultVal: cty.StringVal("nomad-srv")},
		{name: "firewall_name", tpe: "string", defaultVal: cty.StringVal("dev_firewall")},
		{name: "network_name", tpe: "string", defaultVal: cty.StringVal("dev_network")},
		{name: "allow_ips", tpe: "list", defaultVal: cty.TupleVal([]cty.Value{cty.StringVal("85.4.84.201/32")})},
		{name: "server_type", tpe: "string", defaultVal: cty.StringVal("cx21")},
		{name: "location", tpe: "string", defaultVal: cty.StringVal("nbg1")},
	}

	expectedMap := make(map[string]string)
	for _, v := range conf.Variable {
		for _, expected := range vars {
			if expected.name == v.Name {
				expectedMap[expected.name] = expected.name
				assert.Equal(t, expected.tpe, v.Type.Expr.Variables()[0].RootName())
				if expected.defaultVal != cty.NullVal(cty.String) && !strings.Contains(expected.name, "_count") {
					assert.Equal(t, expected.defaultVal, *v.Default)
				}
				if strings.Contains(expected.name, "_count") {
					assert.Equal(t, expected.defaultVal.AsBigFloat().String(), v.Default.AsBigFloat().String())
				}
			}
		}
	}

	assert.Equal(t, len(expectedMap), len(vars))
}

func TestGenerateInventory(t *testing.T) {
	config, err := LoadConfig("testdata/config.yaml")
	assert.NoError(t, err)

	folder := RandString(8)
	config.BaseDir = folder
	err = os.MkdirAll(folder, 0700)
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(filepath.Join(folder))
		assert.NoError(t, err)
	}()

	src := filepath.Join("testdata", "inventory.json")
	dest := filepath.Join(folder, "inventory-output.json")

	bytesRead, err := ioutil.ReadFile(src)
	assert.NoError(t, err)
	fmt.Println(string(bytesRead))

	err = os.WriteFile(filepath.Clean(dest), bytesRead, 0600)
	assert.NoError(t, err)

	err = GenerateInventory(config)
	assert.NoError(t, err)
	bytesRead, err = ioutil.ReadFile(filepath.Join(folder, "inventory"))
	assert.NoError(t, err)
	assert.Equal(t, inventoryResultTest, string(bytesRead))
}
