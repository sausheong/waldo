package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"
	"github.com/joho/godotenv"
	"github.com/sausheong/ishell/v2"
)

var model string
var images []string

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
	gin.SetMode(gin.ReleaseMode)
	go startOllamaServer()
}

func main() {
	shell := ishell.New()

	// display info.
	shell.Println(white("waldo."))
	shell.SetPrompt(getPrompt())

	shell.AddCmd(&ishell.Cmd{
		Name: "shell",
		Help: "run shell commands",
		Func: func(c *ishell.Context) {
			c.Print(cyan("shell> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
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
		Name: "ask",
		Help: "ask waldo",
		Func: func(c *ishell.Context) {
			c.Print(cyan("ask> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				return
			}
			ask(model, line)
			c.Cmd.Func(c)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "search",
		Help: "search the Internet",
		Func: func(c *ishell.Context) {
			c.Print(cyan("search> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				return
			}
			err := search(model, line)
			if err != nil {
				c.Println(red(err))
			}
			c.Cmd.Func(c)
		},
	})

	// ask question about images
	shell.AddCmd(&ishell.Cmd{
		Name: "image",
		Help: "ask questions about an image file",
		Func: func(c *ishell.Context) {
			defer c.SetPrompt(getPrompt())
			if !isImageModel() {
				c.Println(red("Please switch to an image model like llava or Gemini-Pro-Vision or GPT-4-Vision first."))
				return
			}
			if len(images) == 0 {
				c.Print(cyan("image files? "))
				imagesInput := c.ReadLine()
				images = strings.Split(strings.TrimSpace(imagesInput), " ")
			}
			c.Println(yellow(getFilenames(images)))
			c.Print(cyan("image> "))
			line := c.ReadLine()
			if line == "" || line == "exit" {
				return
			}
			if strings.HasPrefix(line, "/clear") {
				images = []string{}
			} else if strings.HasPrefix(line, "/show") {
				if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
					for _, img := range images {
						printInlineImage(img)
					}
				} else {
					fmt.Println(red("showing inline images only available on iTerm2."))
				}
			} else {
				askImage(model, line, images)
			}
			c.Cmd.Func(c)
		},
	})

	// exit walso
	shell.AddCmd(&ishell.Cmd{
		Name: "exit",
		Help: "exit waldo",
		Func: exit,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "switch",
		Help: "switch to a different model",
		Func: func(c *ishell.Context) {
			choices, err := getModels()
			if err != nil {
				log.Println(err)
				c.Println(red("model not switched."))
				return
			}
			choice := c.MultiChoice(choices, cyan("Your current model is ")+yellow(model)+cyan(". Which model to switch to?"))
			model = choices[choice]
			c.SetPrompt(getPrompt())
			c.Println()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "add",
		Help: "add a new model to Waldo",
		Func: func(c *ishell.Context) {
			c.Print(cyan("model name? "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				return
			}
			err := pullModel(line)
			if err != nil {
				c.Println(red(err))
			}
			c.Println()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "info",
		Help: "information about Waldo",
		Func: func(c *ishell.Context) {
			c.Println(yellow("model:"), cyan(model))
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
	return "waldo> "
}

func isImageModel() bool {
	return strings.Contains(model, "llava") || strings.Contains(model, "-vision")
}

func pullModel(name string) error {
	reqJson := `{
	"name": "` + name + `"
}`
	r := bytes.NewReader([]byte(reqJson))
	httpResp, err := http.Post("http://localhost:11435/api/pull", "application/json", r)
	if err != nil {
		fmt.Println("err in calling ollama:", err)
		return err
	}
	decoder := json.NewDecoder(httpResp.Body)
	t0 := time.Now()
	for {
		resp := &PullResponse{}
		decoder.Decode(&resp)
		if resp.Status == "success" {
			elapsed := durafmt.Parse(time.Since(t0)).LimitFirstN(2)
			fmt.Print("\033[2K\r")
			fmt.Printf(cyan("success\n(%s)"), elapsed)
			fmt.Println()
			break
		} else {
			if resp.Status[0:7] == "pulling" {
				fmt.Print("\033[2K\r")
				percentage := float64(resp.Completed) * 100.0 / float64(resp.Total)
				if percentage < 100.0 && percentage != 0.0 {
					fmt.Printf(cyan("downloading ... %.1f%%"), percentage)
				}
			} else {
				fmt.Println(cyan(resp.Status))
			}
		}
	}
	return err
}

func getModels() ([]string, error) {
	models := &Models{}
	httpResp, err := http.Get("http://localhost:11435/api/tags")
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

	results := []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo", "gpt-4-vision", "gemini-pro", "gemini-pro-vision"}
	for _, m := range models.Models {
		results = append(results, m.Name)
	}

	return results, nil
}

func getFilenames(filepaths []string) []string {
	files := []string{}
	for _, path := range filepaths {
		files = append(files, filepath.Base(path))
	}
	return files
}

func printInlineImage(filePath string) {
	bytes, _ := os.ReadFile(filePath)
	b64 := base64.StdEncoding.EncodeToString(bytes)
	fmt.Printf("\x1b]1337;File=inline=1;width=256px:%s\a\n", b64)
}
