package main

import (
	"bufio"
	"bytes"
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
		cmd := exec.Command(args[0], args[1:]...)
		spawnPrograms(cmd)
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

func init() {
	os.Setenv("PROMPT", "->")
}

func main() {
	prompt()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commands := splitOnPipes(scanner.Text())

		var cmds []*exec.Cmd
		for _, command := range commands {
			name, args := parseCommand(command)
			if name == "" {
				continue
			}
			cmd := exec.Command(name, args...)
			cmds = append(cmds, cmd)
		}

		spawnPrograms(cmds...)
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
	commands = strings.Split(line, "|")

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
		fmt.Fprintf(os.Stderr, "osh: command not found: %s\n", name)
		return
	}

	c := exec.Command(cmdFullPath, args...)
	c.Env = os.Environ()
	c.Stdin = placeHolderIn
	c.Stdout = placeHolderOut
	c.Stderr = c.Stdout

	if err = c.Run(); err != nil {
		panic(err)
	}
}

func spawnPrograms(cmds ...*exec.Cmd) {
	stdout, stderr, err := pipeline(cmds)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	if len(stdout) > 0 {
		fmt.Printf("%s", stdout)
	}

	if len(stderr) > 0 {
		fmt.Printf("%s", stderr)
	}
}

func pipeline(cmds []*exec.Cmd) (pipeLineOutput, collectedStandardError []byte, pipeLineError error) {
	if len(cmds) < 1 {
		return nil, nil, nil
	}

	// Collect the output from the command(s)
	var output bytes.Buffer
	var stderr bytes.Buffer

	last := len(cmds) - 1
	for i, cmd := range cmds[:last] {
		var err error
		// Connect each command's stdin to the previous command's stdout
		if cmds[i+1].Stdin, err = cmd.StdoutPipe(); err != nil {
			return nil, nil, err
		}
		// Connect each command's stderr to a buffer
		cmd.Stderr = &stderr
	}

	// Connect the output and error for the last command
	cmds[last].Stdout, cmds[last].Stderr = &output, &stderr

	// Start each command
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Wait for each command to complete
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return output.Bytes(), stderr.Bytes(), err
		}
	}

	// Return the pipeline output and the collected standard error
	return output.Bytes(), stderr.Bytes(), nil
}
