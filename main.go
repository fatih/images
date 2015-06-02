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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	// Create our global configuration and pre-process the argument list to
	// return anything except our global flags. The global flags are passed
	// into the config struct
	config, remainingArgs, err := command.Load(os.Args[1:])
	if err != nil {
		return fmt.Errorf("Error loading global config : %s\n", err)
	}

	fmt.Printf("remainingArgs = %+v\n", remainingArgs)

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
		return fmt.Errorf("Error executing CLI: %s\n", err)
	}

	return nil
}
