package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Destroy(inventory, user string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook destroy.yml -i %s -u %s", inventory, user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err := cmd.Wait()

	return err
}

func Bootstrap(inventory, dcName, user string) error {
	err := Configure(inventory, dcName)
	if err != nil {
		return err
	}
	setup := filepath.Join("config", "setup.yml")
	secrets := filepath.Join("config", "secrets", "secrets.yml")

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook %s -i %s -u %s -e @%s", setup, inventory, user, secrets))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err = cmd.Wait()

	return err
}
