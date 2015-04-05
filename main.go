package main

import (
	"fmt"
	"os"

	"github.com/fatih/images/command"
	"github.com/mitchellh/cli"
)

const (
	Version = "0.0.1"
	Name    = "images"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

func realMain() int {
	c := cli.NewCLI(Name, Version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"list": command.NewList,
		"delete": func() (cli.Command, error) {
			return &cli.MockCommand{SynopsisText: "Delete images"}, nil
		},
		"modify": func() (cli.Command, error) {
			return &cli.MockCommand{SynopsisText: "Modify image properties"}, nil
		},
		"copy": func() (cli.Command, error) {
			return &cli.MockCommand{SynopsisText: "Copy images to different region"}, nil
		},
	}

	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err)
		return 1
	}

	return exitCode
}
