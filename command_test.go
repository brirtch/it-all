package main

import (
	"fmt"
	"testing"
)

func TestParseCommand(t *testing.T) {
	command := " Test test2 \"foo bar\" "
	commandPieces := ParseCommand(command)

	fmt.Println("-" + commandPieces[2] + "-")
	if commandPieces[0] != "Test" || commandPieces[1] != "test2" || commandPieces[2] != "foo bar" {
		t.Error(`ParseCommand(" Test foo \"bar\" ") == false`)
	}
}

func TestSingleWordParseCommand(t *testing.T) {
	command := "time"
	commandPieces := ParseCommand(command)
	fmt.Println(commandPieces[0])

	if commandPieces[0] != "time" {
		t.Error(`ParseCommand("time") == false`)
	}
}
