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
	if d.Provider == "" {
		return `Usage: images delete [options]

  Delete images

Options:

  -provider [name]    Provider to be used to modify images
`
	}

	return Help("delete", d.Provider)
}

func (d *Delete) Run(args []string) int {
	if d.Provider == "" {
		fmt.Print(d.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Print(d.Help())
		return 1
	}

	p, err := Provider(d.Provider, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	deleter, ok := p.(Deleter)
	if !ok {
		err := fmt.Errorf("'%s' doesn't support deleting images", d.Provider)
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
