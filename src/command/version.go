package command

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

type versionCommand struct {
	version string
}

func NewVersion(version string) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &versionCommand{
			version: version,
		}, nil
	}
}

func (v *versionCommand) Help() string     { return "Prints the Images version" }
func (v *versionCommand) Synopsis() string { return "Prints the Images version" }
func (v *versionCommand) Run(args []string) int {
	fmt.Fprintln(os.Stderr, v.version)
	return 0
}
