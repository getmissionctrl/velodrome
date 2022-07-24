package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/relex/aini"
)

//go:embed templates/consul/intention.hcl
var consulIntention string

//go:embed templates/prometheus/prometheus.service
var prometheusService string

//go:embed templates/prometheus/install-prometheus.sh
var prometheusInstall string

//go:embed templates/prometheus/prometheus.yml
var prometheusYml string

//go:embed templates/loki/setup-loki-agent.sh
var lokiDockerAgent string

//go:embed templates/loki/loki.service
var lokiService string

//go:embed templates/loki/loki-config.yml
var lokiConfig string

//go:embed templates/loki/promtail.yml
var promtailConf string

//go:embed templates/loki/promtail.service
var promtailService string

//go:embed templates/ansible/observability.yml
var observabilityAnsible string

//go:embed templates/tempo/setup-tempo.sh
var tempoInstall string

//go:embed templates/tempo/tempo.service
var tempoService string

//go:embed templates/tempo/tempo.yml
var tempoConfig string

//go:embed templates/tempo/tempo-grpc.hcl
var tempoGrpcService string

//go:embed templates/tempo/tempo.hcl
var tempoConsulService string

//go:embed templates/loki/loki-http.hcl
var lokiHttpService string

//go:embed templates/grafana/grafana.hcl
var grafanaHttpService string

//go:embed templates/prometheus/prometheus.hcl
var prometheusConsulService string

type consulServiceConf struct {
	template  string
	hostGroup string
	file      string
	name      string
}

func Observability(inventory, configFile, baseDir, user string) error {
	inv, err := readInventory(inventory)
	if err != nil {
		return err
	}

	secrets := filepath.Join(baseDir, "secrets", "secrets.yml")
	sec, err := getSecrets(baseDir)
	if err != nil {
		return err
	}
	consul := NewConsul(inv, sec, baseDir)

	err = mkObservabilityConfigs(consul, inv, baseDir, user)
	if err != nil {
		return err
	}

	setup := filepath.Join(baseDir, "observability.yml")

	err = runCmd("", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s -e @%s", setup, inventory, user, secrets, configFile), os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func mkObservabilityConfigs(consul Consul, inv *aini.InventoryData, baseDir, user string) error {
	dirs := []string{
		"prometheus", "loki", "grafana", "intentions", "tempo",
	}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(baseDir, dir), 0755)
		if err != nil {
			return err
		}
	}
	toWrite := map[string]string{
		filepath.Join(baseDir, "prometheus", "prometheus.service"):    prometheusService,
		filepath.Join(baseDir, "prometheus", "install-prometheus.sh"): prometheusInstall,
		filepath.Join(baseDir, "observability.yml"):                   observabilityAnsible,
		filepath.Join(baseDir, "loki", "setup-loki-agent.sh"):         lokiDockerAgent,
		filepath.Join(baseDir, "loki", "loki.service"):                lokiService,
		filepath.Join(baseDir, "loki", "promtail.service"):            promtailService,
		filepath.Join(baseDir, "loki", "loki-config.yml"):             lokiConfig,
		filepath.Join(baseDir, "tempo", "tempo.yml"):                  tempoConfig,
		filepath.Join(baseDir, "tempo", "tempo.service"):              tempoService,
		filepath.Join(baseDir, "tempo", "setup-tempo.sh"):             tempoInstall,
		filepath.Join(baseDir, "loki", "promtail.yml"):                promtailConf,
	}
	for k, v := range toWrite {
		err := os.WriteFile(k, []byte(v), 0755)
		if err != nil {
			return err
		}
	}

	clients := getPrivateHosts(inv, "clients")
	consulServers := getPrivateHosts(inv, "consul_servers")
	nomadServers := getPrivateHosts(inv, "nomad_servers")

	tmpl, e := template.New("consul-policies").Parse(prometheusYml)
	if e != nil {
		return e
	}
	var buf bytes.Buffer

	err := tmpl.Execute(&buf, map[string]interface{}{
		"ConsulHosts": append(clients, consulServers...),
		"NomadHosts":  append(clients, nomadServers...),
		"ConsulToken": "{{PROMETHEUS_CONSUL_TOKEN}}",
	})
	if err != nil {
		return err
	}

	output := buf.Bytes()

	err = os.WriteFile(filepath.Join(baseDir, "prometheus", "prometheus.yml"), []byte(output), 0755)
	if err != nil {
		return err
	}

	consulServices := []consulServiceConf{
		{
			template:  tempoGrpcService,
			hostGroup: "tempo",
			file:      filepath.Join(baseDir, "tempo", "tempo-grpc.hcl"),
			name:      "tempo-grpc",
		},
		{
			template:  tempoConsulService,
			hostGroup: "tempo",
			file:      filepath.Join(baseDir, "tempo", "tempo.hcl"),
			name:      "tempo",
		},
		{
			template:  prometheusConsulService,
			hostGroup: "prometheus",
			file:      filepath.Join(baseDir, "prometheus", "prometheus.hcl"),
			name:      "prometheus",
		},
		{
			template:  lokiHttpService,
			hostGroup: "loki",
			file:      filepath.Join(baseDir, "loki", "loki.hcl"),
			name:      "loki",
		},
		{
			template:  grafanaHttpService,
			hostGroup: "grafana",
			file:      filepath.Join(baseDir, "grafana", "grafana.hcl"),
			name:      "grafana",
		},
	}

	for _, service := range consulServices {
		servers := getPrivateHosts(inv, service.hostGroup)
		fmt.Println(servers)
		template := strings.ReplaceAll(service.template, "HOST", servers[0])
		err = os.WriteFile(filepath.Clean(service.file), []byte(template), 0755)
		if err != nil {
			return err
		}
		intention := strings.ReplaceAll(consulIntention, "SRVC", service.name)
		intentionFile := filepath.Join(baseDir, "intentions", fmt.Sprintf("%s.hcl", service.name))
		err = os.WriteFile(intentionFile, []byte(intention), 0755)
		if err != nil {
			return err
		}

		err = consul.RegisterService(service.file)
		if err != nil {
			return err
		}

		err = consul.RegisterIntention(intentionFile)
		if err != nil {
			return err
		}

	}

	return nil
}
