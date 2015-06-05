package awsimages

import (
	"errors"

	"github.com/fatih/images/command/loader"
)

// AwsCommand implements the images various interfaces, such as Fetcher,
// Deleter, Modifier, etc..
type AwsCommand struct {
	*AwsImages
}

// NewCommand returns a new instance of AwsCommand
func NewCommand(args []string) (*AwsCommand, error) {
	var conf struct {
		// just so we can use the Env and TOML loader more efficiently with out
		// any complex hacks
		Aws AwsConfig
	}

	if err := loader.Load(&conf, args); err != nil {
		return nil, err
	}

	awsImages, err := New(&conf.Aws)
	if err != nil {
		return nil, err
	}

	return &AwsCommand{
		AwsImages: awsImages,
	}, nil
}

// List implements the command.Lister interface
func (a *AwsCommand) List(args []string) error {
	images, err := a.ownerImages()
	if err != nil {
		return err
	}

	return images.Print()
}

func (a *AwsCommand) Copy(args []string) error {
	return errors.New("not implemented yet")

}

func (a *AwsCommand) Delete(args []string) error {
	return errors.New("not implemented yet")

}

func (a *AwsCommand) Modify(args []string) error {
	return errors.New("not implemented yet")

}

func (a *AwsCommand) Help(command string) error {
	return errors.New("not implemented yet")
}