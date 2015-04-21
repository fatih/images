package command

import (
	"errors"

	"github.com/fatih/images/provider"
)

func Provider(name string, args []string) (interface{}, error) {
	switch name {
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

func Help(command, name string) string {
	p, err := Provider(name, nil)
	if err != nil {
		return "Provider " + name + " doesn't exists."
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + name
	}

	return h.Help(command)
}
