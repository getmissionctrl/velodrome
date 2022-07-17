package internal

type Config struct {
	User             string `yaml:"sudo_user"`
	DC               string `yaml:"dc_name"`
	Inventory        string `yaml:"inventory"`
	TempoBucket      string `yaml:"tempo_bucket"`
	LokiBucket       string `yaml:"loki_bucket"`
	NetworkInterface string `yaml:"network_interface_name"`
}

func LoadConfig(file string) (*Config, error) {
	return nil, nil
}

type secretsConfig struct {
	ConsulGossipKey        string `yaml:"CONSUL_GOSSIP_KEY"`
	NomadGossipKey         string `yaml:"NOMAD_GOSSIP_KEY"`
	NomadClientConsulToken string `yaml:"NOMAD_CLIENT_CONSUL_TOKEN"`
	NomadServerConsulToken string `yaml:"NOMAD_SERVER_CONSUL_TOKEN"`
	ConsulAgentToken       string `yaml:"CONSUL_AGENT_TOKEN"`
	ConsulBootstrapToken   string `yaml:"CONSUL_BOOTSTRAP_TOKEN"`
}
