package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func bash(command string) (string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	return stdout.String(), stderr.String()
}

func main() {

	artifact := os.Args[1]

	r, _ := regexp.Compile("^dist/(.*).tar.gz$")
	matches := r.FindStringSubmatch(artifact)
	service := matches[1]

	fmt.Printf("Building %s\n", artifact)

	bash(fmt.Sprintf("mkdir -p dist/artifact/%s", service))

	bash(fmt.Sprintf("cp dist/%s dist/artifact/%s/app", service, service))

	bash(fmt.Sprintf("cp deploy/app.service dist/artifact/%s/app.service", service))

	if artifact == "raspberry_client" || artifact == "raspberry_server" {
		bash(fmt.Sprintf("cp dist/libnext.so dist/artifact/%s/libnext.so", service))
	}

	bash(fmt.Sprintf("cd dist/artifact/%s && tar -zcf ../../%s.tar.gz *", service, service))
}
