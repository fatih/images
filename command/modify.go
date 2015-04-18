package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/command/flags"
	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type Modify struct {
	provider string
}

func NewModify(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Modify{
			provider: config.Provider,
		}, nil
	}
}

func (m *Modify) Help() string {
	if m.provider == "" {
		defaultHelp := `Usage: images modify [options]

  Modifies images properties. Each providers sub options are different.

Options:

  -provider                  Provider to be used to modify images
`

		return defaultHelp

	}

	return images.Help("modify", m.provider)
}

func (m *Modify) Run(args []string) int {
	if m.provider == "" {
		fmt.Print(m.Help())
		return 1
	}

	remainingArgs := flags.FilterFlag("provider", args)

	if err := images.Modify(m.provider, remainingArgs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (m *Modify) Synopsis() string {
	return "Modify image properties"
}
