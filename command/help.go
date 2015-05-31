package command

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/mitchellh/cli"
)

// Help is a custom help func for the cli. It's the same as
// cli.BasicHelpFunc but contains our global configuration
func HelpFunc(commands map[string]cli.CommandFactory) string {
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
	cfg := Config{}
	for key, synopsis := range cfg.Help() {
		buf.WriteString(fmt.Sprintf("   -%s    %s\n", key, synopsis))
	}

	return buf.String()
}
