package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
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

type Pokemon struct {
	Id             int
	Name           string
	BaseExperience int `json:"base_experience"`
	Height         int
	Weight         int
	Order          int
	IsDefault      bool `json:"is_default"`
	Stats          []StatData
	Types          []TypeData
}

type StatData struct {
	BaseStat int `json:"base_stat"`
	Effort   int
	Stat     Stat
}

type Stat struct {
	Name string
	Url  string
}

type TypeData struct {
	Slot int
	Type Type
}

type Type struct {
	Id   int
	Name string
}

var commands map[string]cliCommand
var pokemon map[string]Pokemon
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
		},
		"explore": {
			name:        "explore",
			description: "Accepts name of a location and lists Pokemon found there.",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect one of your caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Prints list of all caught pokemon",
			callback:    commandPokedex,
		},
	}
	pokemon = map[string]Pokemon{}

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

func commandCatch(param string, config cmdConfig) error {
	if len(param) == 0 {
		fmt.Println("You must specify a target")
		return nil
	}
	fmt.Printf("\nThrowing a Pokeball at %s...\n", param)

	const apiUrl = "https://pokeapi.co/api/v2/pokemon/"
	url := apiUrl + param
	client := &http.Client{}
	var data []byte

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("You missed, it's almost as if %s wasnt even something you could catch...\n", param)
		return nil
	}

	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var pokemonData Pokemon
	err = json.Unmarshal(data, &pokemonData)
	if err != nil {
		return err
	}
	if rand.Intn(pokemonData.BaseExperience) > pokemonData.BaseExperience-pokemonData.BaseExperience/3 {
		fmt.Println("You caught it!")
		pokemon[param] = pokemonData
	} else {
		fmt.Printf("%s escaped!\n", param)
	}

	return nil
}

func commandInspect(param string, config cmdConfig) error {
	if len(param) == 0 {
		fmt.Println("You must specify a Pokemon to inspect.")
		return nil
	}

	p, ok := pokemon[param]
	if !ok {
		fmt.Printf("You haven't caught %s yet\n", param)
		return nil
	}

	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Height: %d\n", p.Height)
	fmt.Printf("Weight: %d\n", p.Weight)
	fmt.Printf("Stats:\n")
	for _, val := range p.Stats {
		fmt.Printf(" - %s: %d\n", val.Stat.Name, val.BaseStat)
	}
	fmt.Printf("Types:\n")
	for _, val := range p.Types {
		fmt.Printf(" - %s\n", val.Type.Name)
	}

	return nil
}

func commandPokedex(param string, config cmdConfig) error {
	if len(pokemon) == 0 {
		fmt.Println("You havent caught any pokemon")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for _, p := range pokemon {
		fmt.Printf(" - %s\n", p.Name)
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
