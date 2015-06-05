package doimages

import (
	"errors"

	"github.com/fatih/images/command/loader"
)

// DoCommand implements the images various interfaces, such as Fetcher,
// Deleter, Modifier, etc..
type DoCommand struct {
	*DoImages
}

// NewCommand returns a new instance of DoCommand
func NewCommand(args []string) (*DoCommand, error) {
	var conf struct {
		// just so we can use the Env and TOML loader more efficiently with out
		// any complex hacks
		Do DoConfig
	}

	if err := loader.Load(&conf, args); err != nil {
		return nil, err
	}

	doImages, err := New(&conf.Do)
	if err != nil {
		return nil, err
	}

	return &DoCommand{
		DoImages: doImages,
	}, nil
}

// List implements the command.Lister interface
func (d *DoCommand) List(args []string) error {
	images, err := d.UserImages()
	if err != nil {
		return err
	}

	return images.Print()
}

func (d *DoCommand) Copy(args []string) error {
	c := newCopyOptions()
	if err := c.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		c.flagSet.Usage()
		return nil
	}

	if c.ImageID == 0 {
		return errors.New("no image is passed. Use --image")
	}

	return d.CopyImages(c)
}

func (d *DoCommand) Delete(args []string) error {
	return nil
}

// Modify manages the tags of the given images. It can create, override or
// delete tags associated with the given AMI ids.
func (d *DoCommand) Modify(args []string) error {
	return nil
}

// Help prints the help message for the given command
func (d *DoCommand) Help(command string) string {
	var help string
	switch command {
	case "delete":
		help = newDeleteFlags().helpMsg
	case "modify":
		help = newModifyFlags().helpMsg
	case "copy":
		help = newCopyOptions().helpMsg
	case "list":
		help = `Usage: images list --provider do [options]

 List images

Options:
	`
	default:
		return "no help found for command " + command
	}

	global := `
  -token       "..."           DigitalOcean Access Token (env: IMAGES_DO_TOKEN)
`

	help += global
	return help
}
