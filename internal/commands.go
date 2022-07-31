package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
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

func Bootstrap(ctx context.Context, config *Config, configPath string) error {
	inventory := filepath.Join(config.BaseDir, "inventory")
	dcName := config.DC
	user := config.CloudProviderConfig.User
	baseDir := config.BaseDir

	err := GenerateTerraform(config)
	if err != nil {
		return err
	}

	tf, err := InitTf(ctx, filepath.Join(config.BaseDir, "terraform"), os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	token := os.Getenv("HETZNER_TOKEN")
	os.Remove(filepath.Join(config.BaseDir, "inventory-output.json"))

	tf.Apply(ctx, tfexec.Var(fmt.Sprintf("hcloud_token=%s", token)))
	f, err := os.OpenFile(filepath.Join(config.BaseDir, "inventory-output.json"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tf, err = InitTf(ctx, filepath.Join(config.BaseDir, "terraform"), f, os.Stderr)
	if err != nil {
		return err
	}
	tf.Output(ctx)

	err = GenerateInventory(config)
	if err != nil {
		return err
	}
	err = Configure(inventory, baseDir, dcName)
	if err != nil {
		return err
	}
	setup := filepath.Join(baseDir, "setup.yml")
	secrets := filepath.Join(baseDir, "secrets", "secrets.yml")
	fmt.Println("sleeping 10s to ensure all nodes are available..")
	time.Sleep(10 * time.Second)

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
