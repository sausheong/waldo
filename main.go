package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/sausheong/ishell/v2"
)

var model string

var cyan = color.New(color.FgCyan).SprintFunc()
var yellow = color.New(color.FgHiYellow).SprintFunc()
var white = color.New(color.FgWhite, color.Bold).SprintFunc()
var green = color.New(color.FgHiGreen, color.Bold).SprintFunc()
var red = color.New(color.FgRed, color.Bold).SprintFunc()

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	model = os.Getenv("MODEL")
}

func main() {
	shell := ishell.New()

	// display info.
	shell.Println(white("waldo."))
	shell.SetPrompt(getPrompt())

	shell.AddCmd(&ishell.Cmd{
		Name: "sh",
		Help: "run shell commands",
		Func: func(c *ishell.Context) {
			c.Print(cyan("sh> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				c.Println(red("no command, will exit."))
				return
			}
			args := strings.Split(line, " ")
			out, err := run(args[0], args[1:])
			if err != nil {
				c.Println(red(string(out)))
				c.Println(red(err))
				return
			} else {
				if len(out) > 0 {
					c.Println(yellow(string(out)))
				}
			}
			c.Cmd.Func(c)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "search the Internet to answer a question.",
		Func: func(c *ishell.Context) {
			c.Print(cyan("se> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				c.Println(red("no question, will exit."))
				return
			}
			err := search(model, line)
			if err != nil {
				c.Println(red(err))
			}
			c.Cmd.Func(c)
		},
	})

	// exit ghost
	shell.AddCmd(&ishell.Cmd{
		Name: "exit",
		Help: "exit waldo.",
		Func: exit,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "switch",
		Help: "switch to a different model.",
		Func: func(c *ishell.Context) {
			choices, err := getModels()
			if err != nil {
				log.Println(err)
				c.Println(red("model not switched."))
				return
			}
			choice := c.MultiChoice(choices, cyan("Which model to use?"))
			model = choices[choice]
			c.SetPrompt(getPrompt())
			c.Println()
		},
	})

	// start shell
	shell.Run()
	// teardown
	shell.Close()

}

func exit(c *ishell.Context) {
	c.Stop()
}

func getPrompt() string {
	return "wa> "
}

func getModels() ([]string, error) {
	models := &Models{}
	httpResp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Println("err in calling ollama:", err)
		return []string{}, err
	}
	decoder := json.NewDecoder(httpResp.Body)
	err = decoder.Decode(models)
	if err != nil {
		fmt.Println("err in getting models:", err)
		return []string{}, err
	}

	results := []string{}
	for _, m := range models.Models {
		results = append(results, m.Name)
	}

	return results, nil
}
