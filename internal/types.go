package internal

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DC                  string               `yaml:"dc_name"`
	Inventory           string               `yaml:"inventory"`
	BaseDir             string               `yaml:"baseDir"`
	CloudProviderConfig CloudProvider        `yaml:"cloud_provider_config"`
	ObservabilityConfig *ObservabilityConfig `yaml:"observability_config"`
}

type CloudProvider struct {
	User             string `yaml:"sudo_user"`
	NetworkInterface string `yaml:"internal_network_interface_name"`
}

type ObservabilityConfig struct {
	TempoBucket    string `yaml:"tempo_bucket"`
	LokiBucket     string `yaml:"loki_bucket"`
	SingleInstance bool   `yaml:"single_instance"`
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
