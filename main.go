package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Builtin struct {
	Run func(args []string)
}

var BUILTINS = map[string]*Builtin{
	"cd": &Builtin{func(args []string) {
		os.Chdir(args[0])
	}},
	"exit": &Builtin{func(args []string) {
		var code int
		if len(args) == 1 {
			code, _ = strconv.Atoi(args[0])
		}
		os.Exit(code)
	}},
	"exec": &Builtin{func(args []string) {
		spawnProgram(args[0], args[1:])
	}},
	"set": &Builtin{func(args []string) {
		for _, arg := range args {
			keyValuePair := strings.Split(arg, "=")
			if len(keyValuePair) == 2 {
				os.Setenv(keyValuePair[0], keyValuePair[1])
			}
		}
	}},
}

func main() {
	os.Setenv("PROMPT", "->")
	prompt()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name, args := parseCmd(scanner.Text())
		if isBuiltin(name) {
			callBuiltin(name, args)
		} else {
			spawnProgram(name, args)
		}
		prompt()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Reading standard input:", err)
	}
}

func prompt() {
	fmt.Fprintf(os.Stdout, "%s ", os.Getenv("PROMPT"))
}

func parseCmd(text string) (name string, args []string) {
	regexpBySpace := regexp.MustCompile("\\s+")
	cmd := regexpBySpace.Split(text, -1)

	name = cmd[0]
	args = cmd[1:]

	return
}

func isBuiltin(name string) bool {
	_, ok := BUILTINS[name]

	return ok
}

func callBuiltin(name string, args []string) {
	builtin, _ := BUILTINS[name]
	builtin.Run(args)
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
