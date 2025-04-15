package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func main() {
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex >")

		reader.Scan()

		input := reader.Text()

		cleanInput := cleanInput(input)

		cmdMatch := false
		for cmdKey, cmd := range commands() {
			if cmdKey == cleanInput[0] {
				cmd.callback()
				cmdMatch = true
			}
		}

		if !cmdMatch {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range commands() {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func cleanInput(text string) []string {
	var result []string

	words := strings.Split(text, " ")

	for _, w := range words {
		trimmed := strings.Trim(w, " ")
		lower := strings.ToLower(trimmed)

		if len(lower) != 0 {
			result = append(result, lower)
		}
	}

	return result
}

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Display Pokedex help",
			callback:    commandHelp,
		},
	}
}
