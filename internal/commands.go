package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Destroy(inventory, user string) error {
	destroy := filepath.Join("config", "destroy.yml")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s", destroy, inventory, user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err := cmd.Wait()

	return err
}

func Bootstrap(config *Config, configPath string) error {
	inventory := config.Inventory
	dcName := config.DC
	user := config.CloudProviderConfig.User

	err := Configure(inventory, dcName)
	if err != nil {
		return err
	}
	setup := filepath.Join("config", "setup.yml")
	secrets := filepath.Join("config", "secrets", "secrets.yml")

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s -e @%s", setup, inventory, user, secrets, configPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err = cmd.Wait()
	if err != nil {
		return err
	}
	hasBootstrapped, err := BootstrapConsul(inventory)
	if hasBootstrapped {
		fmt.Println("Bootstrapped Consul ACL, re-running Ansible...")
		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s  -e @%s", setup, inventory, user, secrets, configPath))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Start()
		err = cmd.Wait()
	}

	return err
}
