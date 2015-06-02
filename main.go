package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/fatih/images/command"
	"github.com/mitchellh/cli"
)

// The main version number that is being run at the moment. This will be filled
// in by the compiler. A non official release holds the version "dev"
var Version = "dev"

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

func realMain() int {
	// Create our global configuration and pre-process the argument list to
	// return anything except our global flags. The global flags are passed
	// into the config struct
	config, remainingArgs, err := command.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading global config : %s\n", err)
		return 1
	}

	// completely shutdown colors
	if config.NoColor {
		color.NoColor = true
	}

	c := &cli.CLI{
		Name:     "images",
		Version:  Version,
		Args:     remainingArgs,
		HelpFunc: command.HelpFunc,
		Commands: map[string]cli.CommandFactory{
			"list":    command.NewList(config),
			"modify":  command.NewModify(config),
			"delete":  command.NewDelete(config),
			"copy":    command.NewCopy(config),
			"version": command.NewVersion(Version),
		},
	}

	_, err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err)
		return 1
	}

	return 0
}
