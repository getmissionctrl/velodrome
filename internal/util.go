package internal

import (
	"fmt"
	"os/exec"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func HasDependencies() bool {
	dependencies := []string{
		"consul",
		"nomad",
		"ansible-playbook",
		"cfssl",
	}
	missing := []string{}
	for _, v := range dependencies {
		if !commandExists(v) {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		fmt.Println("Dependencies unsatisfied, please install the following applications with your package manager of choice and ensure they are on the PATH:")
	}
	for _, v := range missing {
		fmt.Printf("- %s\n", v)
	}
	return len(missing) == 0

}
