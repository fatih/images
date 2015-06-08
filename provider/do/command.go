package do

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
func NewCommand(args []string) (*DoCommand, []string, error) {
	var conf struct {
		// just so we can use the Env and TOML loader more efficiently with out
		// any complex hacks
		Do DoConfig
	}

	if err := loader.Load(&conf, args); err != nil {
		return nil, nil, err
	}

	doImages, err := New(&conf.Do)
	if err != nil {
		return nil, nil, err
	}

	remainingArgs := loader.ExcludeArgs(&conf, args)

	return &DoCommand{
		DoImages: doImages,
	}, remainingArgs, nil
}

// List implements the command.Lister interface
func (d *DoCommand) List(args []string) error {
	l := newListFlags()
	if err := l.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	images, err := d.UserImages()
	if err != nil {
		return err
	}

	return images.Print(l.output)
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
	df := newDeleteOptions()
	if err := df.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		df.flagSet.Usage()
		return nil
	}

	if len(df.ImageIds) == 0 {
		return errors.New("no images are passed with [--ids]")
	}

	return d.DeleteImages(df)
}

// Modify renames the given images
func (d *DoCommand) Modify(args []string) error {
	r := newRenameOptions()
	if err := r.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		r.flagSet.Usage()
		return nil
	}

	if len(r.ImageIds) == 0 {
		return errors.New("no images are passed with [--ids]")
	}

	if r.Name == "" {
		return errors.New("no name is passed with [--name]")
	}

	return d.RenameImages(r)
}

// Help prints the help message for the given command
func (d *DoCommand) Help(command string) string {
	var help string
	switch command {
	case "delete":
		help = newDeleteOptions().helpMsg
	case "modify":
		help = newRenameOptions().helpMsg
	case "copy":
		help = newCopyOptions().helpMsg
	case "list":
		help = newListFlags().helpMsg
	default:
		return "no help found for command " + command
	}

	global := `
  -token       "..."           DigitalOcean Access Token (env: IMAGES_DO_TOKEN)
`

	help += global
	return help
}
