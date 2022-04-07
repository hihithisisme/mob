package main

import (
	"os/exec"
	"strings"
)

func getCachedChanges() string {
	return silentgit("diff", "--cached", "--stat")
}

func fetch(configuration Configuration) {
	git("fetch", configuration.RemoteName, "--prune")
}

func gitDir() string {
	return silentgit("rev-parse", "--absolute-git-dir")
}

func gitRootDir() string {
	return strings.TrimSuffix(gitDir(), "/.git")
}

func gitUserName() string {
	return silentgitignorefailure("config", "--get", "user.name")
}

func gitUserEmail() string {
	return silentgit("config", "--get", "user.email")
}

func silentgit(args ...string) string {
	commandString, output, err := runCommand("git", args...)

	if err != nil {
		sayGitError(commandString, output, err)
		exit(1)
	}
	return strings.TrimSpace(output)
}

func silentgitignorefailure(args ...string) string {
	_, output, err := runCommand("git", args...)

	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

func gitWithoutEmptyStrings(args ...string) {
	argsWithoutEmptyStrings := deleteEmptyStrings(args)
	git(argsWithoutEmptyStrings...)
}

func git(args ...string) {
	commandString, output, err := runCommand("git", args...)

	if err != nil {
		sayGitError(commandString, output, err)
		exit(1)
	} else {
		sayIndented(commandString)
	}
}

func isGit() bool {
	_, _, err := runCommand("git", "rev-parse")
	return err == nil
}

func gitignorefailure(args ...string) error {
	commandString, output, err := runCommand("git", args...)

	sayIndented(commandString)
	if err != nil {
		sayError(output)
		sayError(err.Error())
	}
	return err
}

func deleteEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func runCommand(name string, args ...string) (string, string, error) {
	command := exec.Command(name, args...)
	if len(workingDir) > 0 {
		command.Dir = workingDir
	}
	commandString := strings.Join(command.Args, " ")
	debugInfo("Running command " + commandString)
	outputBinary, err := command.CombinedOutput()
	output := string(outputBinary)
	debugInfo(output)
	return commandString, output, err
}

func isNothingToCommit() bool {
	output := silentgit("status", "--short")
	return len(output) == 0
}
