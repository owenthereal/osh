package main

import (
	"github.com/bmizerany/assert"
	"os"
	"testing"
)

func TestSplitOnPipes(t *testing.T) {
	commands := splitOnPipes("ps aux | grep foo")
	assert.Equal(t, 2, len(commands))
	assert.Equal(t, "ps aux", commands[0])
	assert.Equal(t, "grep foo", commands[1])
}

func TestParseCommand(t *testing.T) {
	name, args := parseCommand("ls -all")
	assert.Equal(t, "ls", name)
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "-all", args[0])

	name, args = parseCommand("ls     -all")
	assert.Equal(t, "ls", name)
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "-all", args[0])

	os.Setenv("FOO", "BAR")
	name, args = parseCommand("echo $FOO")
	assert.Equal(t, "echo", name)
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "BAR", args[0])
}
