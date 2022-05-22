package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Starting this")

	cmd := exec.Command("/bin/sh", "-c", "ansible-playbook setup.yml -i datacenters/contabo/inventory -u root")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	err := cmd.Wait()

	if err != nil {
		log.Fatal(err)
	}
}
