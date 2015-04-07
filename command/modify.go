package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type Modify struct{}

func NewModify() (cli.Command, error) {
	return &Modify{}, nil
}

func (m *Modify) Help() string {
	return `Usage: images modify [options] 

  Modifies images properties. Each providers sub options are different.

Options:

  -provider                  Provider to be used to modify images
`
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

	if err := images.Modify(provider, remainingArgs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (m *Modify) Synopsis() string {
	return "Modify image properties"
}
