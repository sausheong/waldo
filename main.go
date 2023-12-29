package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"
	"github.com/jmorganca/ollama/format"
	"github.com/jmorganca/ollama/server"
	"github.com/joho/godotenv"
	"github.com/sausheong/ishell/v2"
	"golang.org/x/crypto/ssh"
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
		Name: "ask",
		Help: "ask waldo",
		Func: func(c *ishell.Context) {
			c.Print(cyan("ask> "))
			line := c.ReadLine()
			defer c.SetPrompt(getPrompt())
			if line == "" || line == "exit" {
				c.Println(red("no question, will exit."))
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
			choice := c.MultiChoice(choices, cyan("Which model to use?"))
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
				c.Println(red("no model provided, will exit."))
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
			c.Println(yellow("Waldo is a simple command line application that allows you to ask questions and search the Internet."))
			c.Println(yellow("LLM:"), cyan(model))
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

	results := []string{}
	for _, m := range models.Models {
		results = append(results, m.Name)
	}

	return results, nil
}

// the code below are taken from Ollama to enable the Waldo to
// start an Ollama server within waldo

// start the OllamaServer
func startOllamaServer() error {
	host, port, err := net.SplitHostPort(os.Getenv("OLLAMA_HOST"))
	if err != nil {
		host, port = "127.0.0.1", "11435"
		if ip := net.ParseIP(strings.Trim(os.Getenv("OLLAMA_HOST"), "[]")); ip != nil {
			host = ip.String()
		}
	}

	if err := initializeKeypair(); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	return server.Serve(ln)
}

// initialize the kepair
func initializeKeypair() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	privKeyPath := filepath.Join(home, ".ollama", "id_ed25519")
	pubKeyPath := filepath.Join(home, ".ollama", "id_ed25519.pub")

	_, err = os.Stat(privKeyPath)
	if os.IsNotExist(err) {
		fmt.Printf("Couldn't find '%s'. Generating new private key.\n", privKeyPath)
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		privKeyBytes, err := format.OpenSSHPrivateKey(privKey, "")
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(privKeyPath), 0o755)
		if err != nil {
			return fmt.Errorf("could not create directory %w", err)
		}

		err = os.WriteFile(privKeyPath, pem.EncodeToMemory(privKeyBytes), 0o600)
		if err != nil {
			return err
		}

		sshPrivateKey, err := ssh.NewSignerFromKey(privKey)
		if err != nil {
			return err
		}

		pubKeyData := ssh.MarshalAuthorizedKey(sshPrivateKey.PublicKey())

		err = os.WriteFile(pubKeyPath, pubKeyData, 0o644)
		if err != nil {
			return err
		}

		fmt.Printf("Your new public key is: \n\n%s\n", string(pubKeyData))
	}
	return nil
}
