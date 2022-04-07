package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func openLastModifiedFileIfPresent(configuration Configuration) {
	if !configuration.isOpenCommandGiven() {
		debugInfo("No open command given")
		return
	}

	debugInfo("Try to open last modified file")
	if !lastCommitIsWipCommit(configuration) {
		debugInfo("Last commit isn't a WIP commit.")
		return
	}
	lastCommitMessage := lastCommitMessage()
	split := strings.Split(lastCommitMessage, "lastFile:")
	if len(split) == 1 {
		sayWarning("Couldn't find last modified file in commit message!")
		return
	}
	if len(split) > 2 {
		sayWarning("Could not determine last modified file from commit message, separator was used multiple times!")
		return
	}
	lastModifiedFile := split[1]
	if lastModifiedFile == "" {
		debugInfo("Could not find last modified file in commit message")
		return
	}
	lastModifiedFilePath := gitRootDir() + "/" + lastModifiedFile
	commandname, args := configuration.openCommandFor(lastModifiedFilePath)
	_, err := startCommand(commandname, args...)
	if err != nil {
		sayError(fmt.Sprintf("Couldn't open last modified file on your system (%s)", runtime.GOOS))
		sayError(err.Error())
	}
	debugInfo("Open last modified file: " + lastModifiedFilePath)
}

func sayUntrackedFilesInfo() {
	untrackedFiles := getUntrackedFiles()
	hasUntrackedFiles := len(untrackedFiles) > 0
	if hasUntrackedFiles {
		sayInfo("untracked files present:")
		sayInfoIndented(untrackedFiles)
	}
}

func sayUnstagedChangesInfo() {
	unstagedChanges := getUnstagedChanges()
	hasUnstagedChanges := len(unstagedChanges) > 0
	if hasUnstagedChanges {
		sayInfo("unstaged changes present:")
		sayInfoIndented(unstagedChanges)
	}
}

func getUntrackedFiles() string {
	return silentgit("ls-files", "--others", "--exclude-standard", "--full-name")
}

func getUnstagedChanges() string {
	return silentgit("diff", "--stat")
}

func findStashByName(stashes string, stash string) string {
	lines := strings.Split(stashes, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, stash) {
			return line[:strings.Index(line, ":")]
		}
	}
	return "unknown"
}

func getChangesOfLastCommit() string {
	return silentgit("diff", "HEAD^1", "--stat")
}

func makeWipCommit(configuration Configuration) {
	git("add", "--all")
	commitMessage := createWipCommitMessage(configuration)
	gitWithoutEmptyStrings("commit", "--message", commitMessage, configuration.gitHooksOption())
	sayInfoIndented(getChangesOfLastCommit())
	sayInfoIndented(gitCommitHash())
}

func createWipCommitMessage(configuration Configuration) string {
	commitMessage := configuration.WipCommitMessage

	lastModifiedFilePath := getPathOfLastModifiedFile()
	if lastModifiedFilePath != "" {
		commitMessage += "\n\nlastFile:" + lastModifiedFilePath
	}

	return commitMessage
}

// uses git status --short. To work properly files have to be staged.
func getPathOfLastModifiedFile() string {
	files := getModifiedFiles()
	lastModifiedFilePath := ""
	lastModifiedTime := time.Time{}
	rootDir := gitRootDir()

	debugInfo("Find last modified file")
	if len(files) == 1 {
		lastModifiedFilePath = files[0]
		debugInfo("Just one modified file: " + lastModifiedFilePath)
		return lastModifiedFilePath
	}

	for _, file := range files {
		absoluteFilepath := rootDir + "/" + file
		debugInfo(absoluteFilepath)
		info, err := os.Stat(absoluteFilepath)
		if err != nil {
			sayError("Could not get statistics of file: " + absoluteFilepath)
			sayError(err.Error())
			continue
		}
		modTime := info.ModTime()
		if modTime.After(lastModifiedTime) {
			lastModifiedTime = modTime
			lastModifiedFilePath = file
		}
		debugInfo(modTime.String())
	}
	return lastModifiedFilePath
}

// uses git status --short. To work properly files have to be staged.
func getModifiedFiles() []string {
	debugInfo("Find modified files")
	gitstatus := silentgit("status", "--short")
	lines := strings.Split(gitstatus, "\n")
	files := []string{}
	for _, line := range lines {
		relativeFilepath := ""
		if strings.HasPrefix(line, "M") {
			relativeFilepath = strings.TrimPrefix(line, "M")
		} else if strings.HasPrefix(line, "A") {
			relativeFilepath = strings.TrimPrefix(line, "A")
		} else {
			continue
		}
		relativeFilepath = strings.TrimSpace(relativeFilepath)
		debugInfo(relativeFilepath)
		files = append(files, relativeFilepath)
	}
	return files
}

func (c Configuration) gitHooksOption() string {
	if c.GitHooksEnabled {
		return ""
	} else {
		return "--no-verify"
	}
}

func squashOrNoCommit(configuration Configuration) string {
	if configuration.DoneSquash == Squash {
		return "--squash"
	} else {
		return "--no-commit"
	}
}

func sayLastCommitsList(currentBaseBranch string, currentWipBranch string) {
	commitsBaseWipBranch := currentBaseBranch + ".." + currentWipBranch
	log := silentgit("--no-pager", "log", commitsBaseWipBranch, "--pretty=format:%h %cr <%an>", "--abbrev-commit")
	lines := strings.Split(log, "\n")
	if len(lines) > 5 {
		sayInfo("wip branch '" + currentWipBranch + "' contains " + strconv.Itoa(len(lines)) + " commits. The last 5 were:")
		lines = lines[:5]
	}
	ReverseSlice(lines)
	output := strings.Join(lines, "\n")
	say(output)
}

func hasUncommittedChanges() bool {
	return !isNothingToCommit()
}
