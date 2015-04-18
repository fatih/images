package command

import (
	"errors"

	"github.com/fatih/images/images"
)

func Provider(provider string, args []string) (interface{}, error) {
	switch provider {
	case "aws":
		return images.NewAwsImages(args), nil
	case "digitalocean":
		return nil, errors.New("not supported yet")
	default:
		return nil, errors.New("no such provider available")
	}
}

type Fetcher interface {
	// Fetch fetches the information from the provider
	Fetch(args []string) error

	// Print prints the images to standard output or to something else.
	Print()
}

type Modifier interface {
	Modify(args []string) error
}

type Helper interface {
	Help(command string) string
}

func Help(command, provider string) string {
	p, err := Provider(provider, nil)
	if err != nil {
		return "Provider " + provider + " doesn't exists."
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + provider
	}

	return h.Help(command)
}
