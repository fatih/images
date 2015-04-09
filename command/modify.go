package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type Modify struct {
	provider string
}

func NewModify() (cli.Command, error) {
	// if any provider is passed just get it, we don't care about errors. This
	// is so we can create independent errors
	provider, _, _ := parseFlagValue("provider", os.Args)

	return &Modify{
		provider: provider,
	}, nil
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
	var (
		provider string
	)

	if len(args) == 0 {
		fmt.Print(m.Help())
		return 1
	}

	provider, remainingArgs, err := parseFlagValue("provider", args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	m.provider = provider

	if err := images.Modify(provider, remainingArgs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (m *Modify) Synopsis() string {
	return "Modify image properties"
}
