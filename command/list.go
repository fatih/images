package command

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/cli"
)

type List struct {
	*Config
}

func NewList(config *Config) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &List{
			Config: config,
		}, nil
	}
}

func (l *List) Help() string {
	if len(l.Providers) == 0 {
		return `Usage: images list [options]

  Lists available images for the given providers.

Options:

  -providers "name,..."    Providers to be used to list images
`
	}

	if len(l.Providers) == 1 && l.Providers[0] == "all" {
		return "images: list images for all available providers"
	}

	return Help("list", l.Providers[0])
}

func (l *List) Run(args []string) int {
	if len(l.Providers) == 0 {
		fmt.Println(l.Help())
		return 1
	}

	if flags.Has("help", args) {
		fmt.Println(l.Help())
		return 1
	}

	if len(l.Providers) == 1 && l.Providers[0] == "all" {
		l.Providers = providerList
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	printProvider := func(provider string) error {
		p, err := Provider(provider, args)
		if err != nil {
			if err == errNoProvider {
				return errors.New("Provider '" + provider + "' doesn't exists.")
			}
		}

		lister, ok := p.(Lister)
		if !ok {
			return fmt.Errorf("Provider '%s' doesn't support listing images", provider)
		}

		if err := lister.List(args); err != nil {
			// we don't return here, because Print might display at least
			// successfull results.
			return err
		}

		fmt.Println("")
		return nil
	}

	for _, provider := range l.Providers {
		wg.Add(1)
		go func(provider string) {
			err := printProvider(provider)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
			wg.Done()
		}(provider)
	}

	wg.Wait()

	if multiErrors != nil {
		fmt.Fprintln(os.Stderr, multiErrors.Error())
		return 1
	}

	return 0
}

func (l *List) Synopsis() string {
	return "List available images"
}
