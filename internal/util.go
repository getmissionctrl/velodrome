package internal

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))] //nolint
	}
	return string(b)
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func HasDependencies() bool {
	dependencies := []string{
		"consul",
		"nomad",
		"vault",
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
		fmt.Println("Local dependencies unsatisfied.\n Please install the following applications with your package manager of choice and ensure they are on the PATH:")
	}
	for _, v := range missing {
		fmt.Printf("- %s\n", v)
	}
	return len(missing) == 0
}
