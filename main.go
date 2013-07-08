package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func main() {
	prompt()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := parseCmd(scanner.Text())
		spawnProgram(cmd[0], cmd[1:])
		prompt()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Reading standard input:", err)
	}
}

func prompt() {
	fmt.Fprint(os.Stdout, "-> ")
}

func parseCmd(text string) []string {
	regexpBySpace := regexp.MustCompile("\\s+")
	return regexpBySpace.Split(text, -1)
}

func spawnProgram(name string, args []string) {
	cmdFullPath, err := exec.LookPath(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "osh: command not found: %s", name)
	}

	var stdin, stdout, stderr bytes.Buffer
	c := exec.Command(cmdFullPath, args...)
	c.Stdin = &stdin
	c.Stdout = &stdout
	c.Stderr = &stderr

	err = c.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, stderr.String())
	}

	fmt.Fprint(os.Stdout, stdout.String())
}
