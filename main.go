package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pokecache "github.com/simon-d/pokedex/internal"
)

type cliCommand struct {
	name        string
	description string
	callback    func(param string, config cmdConfig) error
}

type cmdConfig struct {
	Next     string
	Previous string
}

type ApiResponse struct {
	Count    int32
	Next     string
	Previous string
	Results  []Location
}

type Location struct {
	Name string
	Url  string
}

var commands map[string]cliCommand
var nextUrl string
var prevUrl string
var cache *pokecache.Cache

func main() {
	commands = map[string]cliCommand{
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
		"map": {
			name:        "map",
			description: "Display list of 20 location areas. Each subsequent call will display the next 20 locations.",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display list of previous 20 location areas.",
			callback:    commandMapBack,
		}, "explore": {
			name:        "explore",
			description: "Accepts name of a location and lists Pokemon found there.",
			callback:    commandExplore,
		},
	}

	interval, _ := time.ParseDuration("5s")
	cache = pokecache.NewCache(interval)

	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex >")

		reader.Scan()

		input := reader.Text()

		cleanInput := cleanInput(input)

		cmdMatch := false
		for cmdKey, cmd := range commands {
			if cmdKey == cleanInput[0] {
				if len(cleanInput) > 1 {
					cmd.callback(cleanInput[1], cmdConfig{})
				} else {
					cmd.callback("", cmdConfig{})
				}
				cmdMatch = true
			}
		}

		if !cmdMatch {
			fmt.Println("Unknown command")
		}
	}
}

func commandExit(param string, config cmdConfig) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(param string, config cmdConfig) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap(param string, config cmdConfig) error {
	// Need to make a http request
	client := &http.Client{}
	const baseUrl = "https://pokeapi.co/api/v2/location-area"

	var url string
	var data []byte
	if nextUrl == "" {
		url = baseUrl
	} else {
		url = nextUrl
	}

	if entry, exists := cache.Get(url); exists {
		data = entry
	} else {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cache.Add(url, data)
	}

	response := ApiResponse{}
	err := json.Unmarshal(data, &response)

	if err != nil {
		return err
	}

	nextUrl = response.Next
	prevUrl = url
	// fmt.Printf("Next: %s, Prev: %s\n", nextUrl, prevUrl)
	for i := 0; i < len(response.Results); i++ {
		fmt.Println(response.Results[i].Name)
	}

	return nil
}

func commandMapBack(param string, config cmdConfig) error {
	if prevUrl == "" {
		fmt.Printf("you're on the first page\n")
		return nil
	}
	client := &http.Client{}
	var data []byte

	if entry, exists := cache.Get(prevUrl); exists {
		data = entry
	} else {
		req, err := http.NewRequest("GET", prevUrl, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cache.Add(prevUrl, data)
	}

	response := ApiResponse{}
	err := json.Unmarshal(data, &response)

	if err != nil {
		return err
	}

	nextUrl = prevUrl
	prevUrl = response.Previous
	for i := 0; i < len(response.Results); i++ {
		fmt.Println(response.Results[i].Name)
	}

	return nil
}

func commandExplore(param string, config cmdConfig) error {
	fmt.Printf("Exploring %s...\n", param)

	const baseUrl = "https://pokeapi.co/api/v2/location-area/"
	var url = baseUrl + param
	var data []byte

	if entry, exists := cache.Get(url); exists {
		data = entry
	} else {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cache.Add(url, data)
	}

	var response map[string]interface{}
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	// data -> pokemon_encounters -> pokemon* -> name

	// fmt.Print(string(data))
	fmt.Println()

	var locationData interface{}
	err = json.Unmarshal(data, &locationData)
	if err != nil {
		return err
	}

	locations := locationData.(map[string]interface{})

	encounters := locations["pokemon_encounters"].([]interface{})

	var pokemon []string
	for _, val := range encounters {
		e := val.(map[string]interface{})
		p := e["pokemon"].(map[string]interface{})
		pokemon = append(pokemon, p["name"].(string))
	}
	// fmt.Printf("%s\n", encounters)
	for _, p := range pokemon {
		fmt.Println(p)
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
