package docker

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

const (
	DockerCommandName = "docker"
)

func ExecuteDocker(path string, args ...string) (out []string, err error) {
	cmd := exec.Command(DockerCommandName, args...)
	if path != "" {
		cmd.Dir = path
	}
	outByte, err := cmd.CombinedOutput()

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

func IsRunning(name string) (running bool, err error) {
	outLines, err := ExecuteDocker("", "ps")
	alreadyBuild := false
	for _, line := range outLines {
		line = strings.TrimSpace(line)

		if strings.HasSuffix(line, name) {
			alreadyBuild = true
		}
	}
	return alreadyBuild, err
}

func IsExited(name string) (running bool, err error) {
	outLines, err := ExecuteDocker("", "ps", "-a")
	alreadyBuild := false
	for _, line := range outLines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, name) {
			alreadyBuild = strings.Contains(line, "Exited")
		}
	}
	return alreadyBuild, err
}

func Run(path string, repositoryName string) (out []string, err error) {
	return ExecuteDocker(path, "run", "--detach=true", "--name", repositoryName, repositoryName)
}
