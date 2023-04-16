package main

import "strings"

type Command struct {
	Command string `json:"command"`
}

func ParseCommand(command string) []string {
	command = strings.Trim(command, " ")
	var commandPieces []string
	inQuotes := false
	currentPiece := ""
	for pos, char := range command {
		if char == '"' {
			inQuotes = !inQuotes
		}
		if char == ' ' && !inQuotes {
			commandPieces = append(commandPieces, currentPiece)
			currentPiece = ""
		} else {
			if char != '"' {
				currentPiece += string(char)
			}
		}
		if pos == len(command)-1 {
			commandPieces = append(commandPieces, currentPiece)
		}
	}

	return commandPieces
}
