package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

//go:embed templates/prometheus/prometheus.hcl
var prometheusConsulService string

type consulServiceConf struct {
	template  string
	hostGroup string
	file      string
	name      string
}

func Observability(inventory, user string) error {

	err := mkObservabilityConfigs(inventory, user)
	if err != nil {
		return err
	}

	setup := filepath.Join("config", "observability.yml")
	secrets := filepath.Join("config", "secrets", "secrets.yml")

	err = runCmd("", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s", setup, inventory, user, secrets), os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func mkObservabilityConfigs(inventory, user string) error {
	err := os.MkdirAll(filepath.Join("config", "prometheus"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join("config", "loki"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join("config", "grafana"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join("config", "intentions"), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join("config", "tempo"), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "prometheus", "prometheus.service"), []byte(prometheusService), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "prometheus", "install-prometheus.sh"), []byte(prometheusInstall), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "observability.yml"), []byte(observabilityAnsible), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "loki", "setup-loki-agent.sh"), []byte(lokiDockerAgent), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "loki", "loki.service"), []byte(lokiService), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "loki", "promtail.service"), []byte(promtailService), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "loki", "loki-config.yml"), []byte(lokiConfig), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "tempo", "tempo.yml"), []byte(tempoConfig), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "tempo", "tempo.service"), []byte(tempoService), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "tempo", "setup-tempo.sh"), []byte(tempoInstall), 0755)
	if err != nil {
		return err
	}

	inv, err := readInventory(inventory)
	if err != nil {
		return err
	}
	clients := getPrivateHosts(inv, "clients")
	consulServers := getPrivateHosts(inv, "consul_servers")
	nomadServers := getPrivateHosts(inv, "nomad_servers")

	tmpl, e := template.New("consul-policies").Parse(prometheusYml)
	if e != nil {
		return e
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string][]string{
		"ConsulHosts": append(clients, consulServers...),
		"NomadHosts":  append(clients, nomadServers...),
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join("config", "loki", "promtail.yml"), []byte(promtailConf), 0755)
	if err != nil {
		return err
	}

	output := buf.Bytes()

	err = os.WriteFile(filepath.Join("config", "prometheus", "prometheus.yml"), []byte(output), 0755)

	secrets, err := getSecrets()
	if err != nil {
		return err
	}

	consulServices := []consulServiceConf{
		{
			template:  tempoGrpcService,
			hostGroup: "tempo",
			file:      filepath.Join("config", "tempo", "tempo-grpc.hcl"),
			name:      "tempo-grpc",
		},
		{
			template:  tempoConsulService,
			hostGroup: "tempo",
			file:      filepath.Join("config", "tempo", "tempo.hcl"),
			name:      "tempo",
		},
		{
			template:  prometheusConsulService,
			hostGroup: "prometheus",
			file:      filepath.Join("config", "prometheus", "prometheus.hcl"),
			name:      "prometheus",
		},
		{
			template:  lokiHttpService,
			hostGroup: "loki",
			file:      filepath.Join("config", "loki", "loki.hcl"),
			name:      "loki",
		},
	}

	for _, service := range consulServices {
		servers := getPrivateHosts(inv, service.hostGroup)
		template := strings.ReplaceAll(service.template, "HOST", servers[0])
		err = os.WriteFile(filepath.Clean(service.file), []byte(template), 0755)
		if err != nil {
			return err
		}
		intention := strings.ReplaceAll(consulIntention, "SRVC", service.name)
		intentionFile := filepath.Join("config", "intentions", fmt.Sprintf("%s.hcl", service.name))
		err = os.WriteFile(intentionFile, []byte(intention), 0755)
		if err != nil {
			return err
		}

		err = registerConsul(inv, secrets, service.file)
		if err != nil {
			return err
		}

		err = registerIntention(inv, secrets, intentionFile)
		if err != nil {
			return err
		}

	}

	return nil
}
