package command

import (
	"fmt"
	"os"

	"github.com/fatih/flags"
	"github.com/mitchellh/cli"
)

type Delete struct {
	*Config
}

func NewDelete(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &Delete{
			Config: config,
		}, nil
	}
}

func (d *Delete) Help() string {
	if len(d.Providers) != 1 {
		return `Usage: images delete [options]

  Delete images

Options:

  -providers [name]    Provider to be used to modify images
`
	}

	return Help("delete", d.Providers[0])
}

func (d *Delete) Run(args []string) int {
	if len(d.Providers) != 1 {
		fmt.Print(d.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(d.Help())
		return 1
	}

	provider := d.Providers[0]
	if provider == "all" {
		fmt.Fprintln(os.Stderr, "Delete doesn't support multiple providers")
		return 1
	}

	p, err := Provider(provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	deleter, ok := p.(Deleter)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support deleting images", provider)
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Don't ask for question if --force is enabled
	if !d.Force {
		response, err := d.Ui.Ask("Do you really want to delete? (Type 'yes' to continue):")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}

		if response != "yes" {
			d.Ui.Output("Delete cancelled.")
			return 0
		}
	}

	if err := deleter.Delete(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (d *Delete) Synopsis() string {
	return "Delete available images"
}
