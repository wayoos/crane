package docker

import (
	"bufio"
	"bytes"
	"os/exec"
)

const (
	DockerCommandName = "docker"
)

func ExecuteDocker(path string, args ...string) (out []string, err error) {
	cmd := exec.Command(DockerCommandName, args...)
	cmd.Dir = path
	outByte, err := cmd.CombinedOutput()
	//	if err != nil {
	//		return
	//	}

	reader := bytes.NewReader(outByte)
	scanner := bufio.NewScanner(reader)

	outLines := []string{}

	for scanner.Scan() {
		outLines = append(outLines, scanner.Text())
	}

	return outLines, err
}

func Build(path string, repositoryName string) (out []string, err error) {
	return ExecuteDocker(path, "build", "-t", repositoryName, ".")
}

func Run(path string, repositoryName string) (out []string, err error) {
	return ExecuteDocker(path, "run", "-name", repositoryName, repositoryName)
}
