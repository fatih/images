package aws

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
	c := newCopyOptions()
	if err := c.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		c.flagSet.Usage()
		return nil
	}

	if c.ImageID == "" {
		return errors.New("no image is passed. Use --image")
	}

	return a.CopyImages(c)
}

func (a *AwsCommand) Delete(args []string) error {
	d := newDeleteOptions()
	if err := d.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		d.flagSet.Usage()
		return nil
	}

	if len(d.ImageIds) == 0 {
		return errors.New("no images are passed with [--ids]")
	}

	return a.DeleteImages(d)
}

// Modify manages the tags of the given images. It can create, override or
// delete tags associated with the given AMI ids.
func (a *AwsCommand) Modify(args []string) error {
	m := newModifyFlags()
	if err := m.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		m.flagSet.Usage()
		return nil
	}

	if len(m.imageIds) == 0 {
		return errors.New("no images are passed with [--ids]")
	}

	if m.createTags != "" && m.deleteTags != "" {
		return errors.New("not allowed to be used together: [--create-tags,--delete-tags]")
	}

	if m.createTags != "" {
		return a.CreateTags(m.createTags, m.dryRun, m.imageIds...)
	}

	if m.deleteTags != "" {
		return a.DeleteTags(m.deleteTags, m.dryRun, m.imageIds...)
	}

	return nil
}

// Help prints the help message for the given command
func (a *AwsCommand) Help(command string) string {
	var help string

	global := `
  -access-key      "..."       AWS Access Key (env: IMAGES_AWS_ACCESS_KEY)
  -secret-key      "..."       AWS Secret Key (env: IMAGES_AWS_SECRET_KEY)
  -regions         "..."       AWS Regions (env: IMAGES_AWS_REGION)
  -regions-exclude "..."       AWS Regions to be excluded (env: IMAGES_AWS_REGION_EXCLUDE)
`
	switch command {
	case "modify":
		help = newModifyFlags().helpMsg
	case "delete":
		help = newDeleteOptions().helpMsg
	case "list":
		help = `Usage: images list --provider aws [options]

 List AMI properties.

Options:
	`
	case "copy":
		help = newCopyOptions().helpMsg
	default:
		return "no help found for command " + command
	}

	help += global
	return help
}
