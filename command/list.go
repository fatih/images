package command

import (
	"fmt"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type List struct{}

func NewList() (cli.Command, error) {
	return &List{}, nil
}

func (l *List) Help() string {
	return `Usage: images list [options]

  Lists available images for the given provider.

Options:

  -provider                  Provider to be used to modify images
`
}

func (l *List) Run(args []string) int {
	var (
		provider string
	)

	if len(args) == 0 {
		fmt.Print(l.Help())
		return 1
	}

	provider, err := providerFromEnvOrFlag(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if err := images.List(provider); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
