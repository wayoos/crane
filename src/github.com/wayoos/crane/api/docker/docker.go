package docker

import (
	"bufio"
	"bytes"
	"github.com/wayoos/crane/api/domain"
	"os/exec"
	"strings"
	"syscall"
)

const (
	DockerCommandName = "docker"
)

func ExecuteDocker(path string, args ...string) (out []string, appErr *domain.AppError) {
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

	if cmd.ProcessState.Success() {
		return outLines, nil
	} else {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return outLines, &domain.AppError{err, "Docker command error.", status.ExitStatus()}
	}

}

func Build(path string, repositoryName string) (out []string, appErr *domain.AppError) {
	return ExecuteDocker(path, "build", "-t", repositoryName, ".")
}

func IsRunning(name string) (running bool, appErr *domain.AppError) {
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

func IsExited(name string) (running bool, appErr *domain.AppError) {
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

func Run(path string, repositoryName string) (out []string, err *domain.AppError) {
	return ExecuteDocker(path, "run", "--detach=true", "--name", repositoryName, repositoryName)
}

func Start(container string) (out []string, err *domain.AppError) {
	return ExecuteDocker("", "start", container)
}

func Stop(container string) (out []string, err *domain.AppError) {
	return ExecuteDocker("", "stop", container)
}

func RemoveContainer(container string) (out []string, err *domain.AppError) {
	return ExecuteDocker("", "rm", container)
}
