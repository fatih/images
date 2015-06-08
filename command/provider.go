package command

import (
	"errors"
	"fmt"

	"github.com/fatih/images/provider/aws"
	"github.com/fatih/images/provider/do"
	"github.com/fatih/images/provider/gce"
)

var (
	errNoProvider = errors.New("no such provider available")

	// providerList is used when the provider name is set as "all"
	providerList = []string{
		"aws",
		"do",
		"gce",
	}
)

// Provider returns the provider with the given name and the filtered remaining
// arguments. Each provider is responsible of how the remaining arguments are
// returned.
func Provider(name string, args []string) (interface{}, []string, error) {
	switch name {
	case "aws":
		return aws.NewCommand(args)
	case "do":
		return do.NewCommand(args)
	case "gce":
		return gce.NewCommand(args)
	default:
		return nil, nil, errNoProvider
	}
}

// Lister lists and prints the images
type Lister interface {
	List(args []string) error
}

// Copyier copyies the image.
type Copyier interface {
	Copy(args []string) error
}

// Deleter delets images
type Deleter interface {
	Delete(args []string) error
}

// Modifier modifies image attributes
type Modifier interface {
	Modify(args []string) error
}

// Helper returns the help message
type Helper interface {
	Help(command string) string
}

func Help(command, name string) string {
	p, _, err := Provider(name, nil)
	if err != nil {
		if err == errNoProvider {
			return "Provider '" + name + "' doesn't exists."
		}

		// provider exists but failed, print the help for it
		fmt.Println(err, "\n")
	}

	h, ok := p.(Helper)
	if !ok {
		return "No help context available for " + name
	}

	return h.Help(command)
}
