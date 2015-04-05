package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/images/images"
	"github.com/mitchellh/cli"
)

type List struct {
	flagSet  *flag.FlagSet
	provider string
}

func NewList() (cli.Command, error) {
	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	l := &List{flagSet: flagSet}

	flagSet.StringVar(&l.provider, "provider", "", "cloud provider to list images")
	flagSet.Usage = func() {
		l.Help()
	}

	return l, nil
}

func (l *List) Help() string {
	help := "Usage: images list [options]\n\n"
	help += l.Synopsis() + "\n\n"
	l.flagSet.VisitAll(func(fl *flag.Flag) {
		format := "  -%s=%s\t %s\n"
		help += fmt.Sprintf(format, fl.Name, fl.DefValue, fl.Usage)
	})

	help += "\n"
	return help
}

func (l *List) Run(args []string) int {
	if err := l.flagSet.Parse(args); err != nil {
		return 1
	}

	if l.flagSet.NFlag() == 0 {
		fmt.Print(l.Help())
		return 1
	}

	if err := images.Run(&images.Config{
		Provider: l.provider,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
