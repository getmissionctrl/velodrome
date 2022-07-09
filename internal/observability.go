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

	inv, err := readInventory(inventory)
	if err != nil {
		return err
	}
	clients := getPrivateHosts(inv, "clients")
	consulServers := getPrivateHosts(inv, "consul_servers")
	nomadServers := getPrivateHosts(inv, "nomad_servers")
	lokiServers := getPrivateHosts(inv, "loki")

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

	promtailConf = strings.ReplaceAll(promtailConf, "[HOST]", lokiServers[0])

	err = os.WriteFile(filepath.Join("config", "loki", "promtail.yml"), []byte(promtailConf), 0755)
	if err != nil {
		return err
	}

	output := buf.Bytes()

	err = os.WriteFile(filepath.Join("config", "prometheus", "prometheus.yml"), []byte(output), 0755)

	return nil
}
