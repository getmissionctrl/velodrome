package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Destroy(inventory, baseDir, user string) error {
	destroy := filepath.Join(baseDir, "destroy.yml")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s", destroy, inventory, user)) //nolint
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()

	return err
}

func Bootstrap(config *Config, configPath string) error {
	inventory := config.Inventory
	dcName := config.DC
	user := config.CloudProviderConfig.User
	baseDir := config.BaseDir

	fmt.Println(config.ObservabilityConfig)

	err := Configure(inventory, baseDir, dcName)
	if err != nil {
		return err
	}
	setup := filepath.Join(baseDir, "setup.yml")
	secrets := filepath.Join(baseDir, "secrets", "secrets.yml")

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s -e @%s", setup, inventory, user, secrets, configPath)) //nolint
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	inv, err := readInventory(inventory)
	if err != nil {
		return err
	}
	sec, err := getSecrets(baseDir)
	if err != nil {
		return err
	}
	consul := NewConsul(inv, sec, baseDir)
	hasBootstrapped, err := BootstrapConsul(consul, inv, baseDir)
	if err != nil {
		return err
	}
	if hasBootstrapped {
		fmt.Println("Bootstrapped Consul ACL, re-running Ansible...")
		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s  -e @%s", setup, inventory, user, secrets, configPath)) //nolint
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
	}

	return Observability(inventory, configPath, baseDir, user)

}
