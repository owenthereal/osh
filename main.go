package main

import (
	"bufio"
	"fmt"
	"io"
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
		spawnProgram(args[0], args[1:], os.Stdin, os.Stdout)
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
		commands := splitOnPipes(scanner.Text())
		var placeHolderIn io.ReadCloser = os.Stdin
		var placeHolderOut io.WriteCloser = os.Stdout
		var pipeReader *io.PipeReader

		for i, command := range commands {
			name, args := parseCommand(command)
			if isBuiltin(name) {
				callBuiltin(name, args)
			} else {
				if i+1 < len(commands) {
					pipeReader, placeHolderOut = io.Pipe()
				} else {
					placeHolderOut = os.Stdout
				}

				spawnProgram(name, args, placeHolderOut, placeHolderIn)

				if placeHolderOut != os.Stdout {
					placeHolderOut.Close()
				}

				if placeHolderIn != os.Stdin {
					placeHolderIn.Close()
				}

				placeHolderIn = pipeReader
			}
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

func splitOnPipes(line string) (commands []string) {
	pipesRegexp := regexp.MustCompile("([^\"'|]+)|[\"']([^\"']+)[\"']")
	if pipesRegexp.MatchString(line) {
		commands = pipesRegexp.FindAllString(line, -1)
	} else {
		commands = append(commands, line)
	}

	for i, command := range commands {
		commands[i] = strings.TrimSpace(command)
	}

	return
}

func parseCommand(line string) (name string, args []string) {
	regexpBySpace := regexp.MustCompile("\\s+")
	cmd := regexpBySpace.Split(line, -1)

	name = cmd[0]
	// expand environment variables
	// somehow os/exec.Command.Run() doesn't expand automatically
	envVarRegexp := regexp.MustCompile("^\\$(.+)$")
	for _, arg := range cmd[1:] {
		if envVarRegexp.MatchString(arg) {
			match := envVarRegexp.FindStringSubmatch(arg)
			arg = os.Getenv(match[1])
		}

		args = append(args, arg)
	}

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

func spawnProgram(name string, args []string, placeHolderOut io.WriteCloser, placeHolderIn io.ReadCloser) {
	cmdFullPath, err := exec.LookPath(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "osh: command not found: %s", name)
	}

	c := exec.Command(cmdFullPath, args...)
	c.Env = os.Environ()

	c.Stdin = placeHolderIn
	c.Stdout = placeHolderOut
	c.Stderr = c.Stdout

	err = c.Run()
	if err != nil {
		//fmt.Fprintln(os.Stderr, stderr.String())
		panic(err)
	}
}
