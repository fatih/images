package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
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

	config, err := command.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading global config : %s\n", err)
		return 1
	}

	// completely shutdown colors
	if config.NoColor {
		color.NoColor = true
	}

	c.Commands = map[string]cli.CommandFactory{
		"list":   command.NewList(config),
		"modify": command.NewModify(config),
		"delete": command.NewDelete(config),
		"copy":   command.NewCopy(config),
	}

	_, err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err)
		return 1
	}

	return 0
}
