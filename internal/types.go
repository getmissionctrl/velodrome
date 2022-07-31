package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DC                  string              `yaml:"dc_name"`
	BaseDir             string              `yaml:"baseDir"`
	CloudProviderConfig CloudProvider       `yaml:"cloud_provider_config"`
	ClusterConfig       ClusterConfig       `yaml:"cluster_config"`
	ObservabilityConfig ObservabilityConfig `yaml:"observability_config"`
}

type ClusterConfig struct {
	Servers               int  `yaml:"servers"`
	Clients               int  `yaml:"clients"`
	SeparateConsulServers bool `yaml:"separate_consul_servers"`
}

type CloudProvider struct {
	User             string                 `yaml:"sudo_user"`
	NetworkInterface string                 `yaml:"internal_network_interface_name"`
	Provider         string                 `yaml:"provider"`
	ProviderSettings map[string]interface{} `yaml:"provider_settings"`
}

type ObservabilityConfig struct {
	TempoBucket   string `yaml:"tempo_bucket"`
	LokiBucket    string `yaml:"loki_bucket"`
	MultiInstance bool   `yaml:"multi_instance"`
}

type HetznerResourceNames struct {
	BaseServerName string `yaml:"base_server_name"`
	FirewallName   string `yaml:"firewall_name"`
	NetworkName    string `yaml:"network_name"`
}

type HetznerSettings struct {
	Location      string               `yaml:"location"`
	SSHKeys       []string             `yaml:"ssh_keys"`
	AllowedIPs    []string             `yaml:"allowed_ips"`
	ServerType    string               `yaml:"server_type"`
	ResourceNames HetznerResourceNames `yaml:"resource_names"`
}

type secretsConfig struct {
	ConsulGossipKey        string `yaml:"CONSUL_GOSSIP_KEY"`
	NomadGossipKey         string `yaml:"NOMAD_GOSSIP_KEY"`
	NomadClientConsulToken string `yaml:"NOMAD_CLIENT_CONSUL_TOKEN"`
	NomadServerConsulToken string `yaml:"NOMAD_SERVER_CONSUL_TOKEN"`
	ConsulAgentToken       string `yaml:"CONSUL_AGENT_TOKEN"`
	ConsulBootstrapToken   string `yaml:"CONSUL_BOOTSTRAP_TOKEN"`
	PrometheusConsulToken  string `yaml:"PROMETHEUS_CONSUL_TOKEN"`
	S3Endpoint             string `yaml:"s3_endpoint"`
	S3AccessKey            string `yaml:"s3_access_key"`
	S3SecretKey            string `yaml:"s3_secret_key"`
}

type TFVarsConfig struct {
	ClusterConfig  ClusterConfig
	ProviderConfig interface{}
}

func LoadConfig(file string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadTFVarsConfig(config Config) (*TFVarsConfig, error) {
	var providerConfig interface{}
	if config.CloudProviderConfig.Provider == "hetzner" {
		var hetznerConfig HetznerSettings
		bytes, err := yaml.Marshal(config.CloudProviderConfig.ProviderSettings)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(bytes, &hetznerConfig)
		if err != nil {
			return nil, err
		}
		providerConfig = hetznerConfig
	}

	return &TFVarsConfig{
		ClusterConfig:  config.ClusterConfig,
		ProviderConfig: providerConfig,
	}, nil
}

func LoadTFExecVars(config *Config) *tfexec.VarOption {

	token := os.Getenv("HETZNER_TOKEN")
	return tfexec.Var(fmt.Sprintf("hcloud_token=%s", token))
}
