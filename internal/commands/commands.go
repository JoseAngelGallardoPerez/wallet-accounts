package commands

import (
	"fmt"
	"log"
	"net/url"

	"go.uber.org/dig"
)

type command struct {
	Name        string
	Usage       string
	Description string
	Handler     handler
}

type handler func(c *dig.Container, args url.Values)

var commands = map[string]command{
	"help": {
		Name:        "help",
		Usage:       "help",
		Description: "Display available commands",
		Handler:     func(c *dig.Container, args url.Values) {},
	},
	executeScheduledTransaction.Name: executeScheduledTransaction,
}

func Process(cmd string, c *dig.Container) {
	u, err := url.Parse(cmd)
	if err != nil {
		log.Fatal("unable to parse command: " + cmd + "; command must be valid url format e.g. cmd?param=value")
	}

	if command, exist := commands[u.Path]; exist {
		if command.Name == "help" {
			for _, v := range commands {
				fmt.Printf("%s:\nUsage:\t%s\nDescription:\t%s\n\n", v.Name, v.Usage, v.Description)
			}
			return
		}
		q, _ := url.ParseQuery(u.RawQuery)
		command.Handler(c, q)
		return
	}

	log.Fatal("command not found: " + u.Path)
}
