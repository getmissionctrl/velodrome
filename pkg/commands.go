package pkg

import (
	"fmt"
	"os"
	"os/exec"
)

func Destroy(inventory, user string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("ansible-playbook destroy.yml -i %s -u %s", inventory, user))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err := cmd.Wait()

	return err
}
