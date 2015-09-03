package command

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"text/tabwriter"

	"github.com/mitchellh/cli"
)

// Help is a custom help func for the cli. It's the same as
// cli.BasicHelpFunc but contains our global configuration
func HelpFunc(commands map[string]cli.CommandFactory) string {
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 10, 8, 0, '\t', 0)

	fmt.Fprintf(w, "usage: images [--version] [--help] <command> [<args>]\n\n")
	fmt.Fprintf(w, "Available commands are:\n")

	// Get the list of keys so we can sort them
	keys := make([]string, 0, len(commands))
	for key := range commands {
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

		fmt.Fprintf(w, "    %s\t%s\n", key, command.Synopsis())
	}

	fmt.Fprintf(w, "\nAvailable global flags are:\n")

	cfg := Config{}
	helps := cfg.Help()
	globals := make([]string, 0, len(helps))
	for key := range helps {
		globals = append(globals, key)
	}
	sort.Strings(globals)

	for _, flag := range globals {
		fmt.Fprintf(w, "   -%s\t%s\n", flag, helps[flag])
	}

	w.Flush()
	return buf.String()
}
