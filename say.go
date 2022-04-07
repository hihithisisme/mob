package main

import (
	"fmt"
	"strings"
)

func sayError(text string) {
	sayWithPrefix(text, "ERROR ")
}

func debugInfo(text string) {
	if Debug {
		sayWithPrefix(text, "DEBUG ")
	}
}

func sayIndented(text string) {
	sayWithPrefix(text, "  ")
}

func sayFix(instruction string, command string) {
	sayWithPrefix(instruction, "👉 ")
	sayEmptyLine()
	sayIndented(command)
	sayEmptyLine()
}

func sayNext(instruction string, command string) {
	sayWithPrefix(instruction, "👉 ")
	sayEmptyLine()
	sayIndented(command)
	sayEmptyLine()
}

func sayInfo(text string) {
	sayWithPrefix(text, "> ")
}

func sayInfoIndented(text string) {
	sayWithPrefix(text, "    ")
}

func sayWarning(text string) {
	sayWithPrefix(text, "⚠ ")
}

func sayWithPrefix(s string, prefix string) {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := 0; i < len(lines); i++ {
		printToConsole(prefix + strings.TrimSpace(lines[i]) + "\n")
	}
}

func say(s string) {
	if len(s) == 0 {
		return
	}
	printToConsole(strings.TrimRight(s, " \r\n\t\v\f") + "\n")
}

func sayEmptyLine() {
	printToConsole("\n")
}

var printToConsole = func(message string) {
	fmt.Print(message)
}

func sayGitError(commandString string, output string, err error) {
	if !isGit() {
		sayError("expecting the current working directory to be a git repository.")
	} else {
		sayError(commandString)
		sayError(output)
		sayError(err.Error())
	}
}
