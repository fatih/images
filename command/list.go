package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type List struct {
	provider string
}

func NewList(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &List{
			provider: config.Provider,
		}, nil
	}
}

func (l *List) Help() string {
	return `Usage: images list [options]

  Lists available images for the given provider.

Options:

  -provider [name]    Provider to be used to modify images
`
}

func (l *List) Run(args []string) int {
	if l.provider == "" {
		fmt.Print(l.Help())
		return 1
	}

	if err := images.List(l.provider); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
