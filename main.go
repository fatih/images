package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/fatih/images/command"
	"github.com/mitchellh/cli"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

func realMain() int {
	c := cli.NewCLI("images", Version)
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
		"list":    command.NewList(config),
		"modify":  command.NewModify(config),
		"delete":  command.NewDelete(config),
		"copy":    command.NewCopy(config),
		"version": command.NewVersion(Version),
	}

	c.HelpFunc = imagesHelp

	_, err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err)
		return 1
	}

	return 0
}

// imagesHelp is a custom help func for the cli. It's the same as
// cli.BasicHelpFunc but contains our global configuration
func imagesHelp(commands map[string]cli.CommandFactory) string {
	var buf bytes.Buffer
	buf.WriteString("usage: images [--version] [--help] <command> [<args>]\n\n")
	buf.WriteString("Available commands are:\n")

	// Get the list of keys so we can sort them, and also get the maximum
	// key length so they can be aligned properly.
	keys := make([]string, 0, len(commands))
	maxKeyLen := 0
	for key, _ := range commands {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}

		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		commandFunc, ok := commands[key]
		if !ok {
			// This should never happen since we JUST built the list of
			// keys.
			panic("command not found: " + key)
		}

		command, err := commandFunc()
		if err != nil {
			log.Printf("[ERR] cli: Command '%s' failed to load: %s",
				key, err)
			continue
		}

		// +2 just comes from the global configuration name, which is longer
		// than the commands above, hacky I know but for now it does the work
		key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)+2))
		buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
	}

	buf.WriteString("\nAvailable global flags are:\n")
	cfg := command.Config{}
	for key, synopsis := range cfg.Help() {
		buf.WriteString(fmt.Sprintf("   -%s    %s\n", key, synopsis))
	}

	return buf.String()
}
