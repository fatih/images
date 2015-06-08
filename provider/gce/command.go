package gce

import (
	"errors"

	"github.com/fatih/images/command/loader"
)

// GceCommand implements the images various interfaces, such as Fetcher,
// Deleter, Modifier, etc..
type GceCommand struct {
	*GceImages
}

// NewCommand returns a new instance of GceImages
func NewCommand(args []string) (*GceCommand, []string, error) {
	var conf struct {
		// just so we can use the Env and TOML loader more efficiently with out
		// any complex hacks
		Gce GCEConfig
	}

	if err := loader.Load(&conf, args); err != nil {
		return nil, nil, err
	}

	gceImages, err := New(&conf.Gce)
	if err != nil {
		return nil, nil, err
	}

	remainingArgs := loader.ExcludeArgs(&conf, args)
	return &GceCommand{
		GceImages: gceImages,
	}, remainingArgs, nil
}

// List implements the command.Lister interface
func (g *GceCommand) List(args []string) error {
	l := newListFlags()
	if err := l.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	images, err := g.ProjectImages()
	if err != nil {
		return err
	}

	return images.Print(l.output)
}

func (g *GceCommand) Delete(args []string) error {
	df := newDeleteOptions()
	if err := df.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		df.flagSet.Usage()
		return nil
	}

	if len(df.Names) == 0 {
		return errors.New("no images are passed with [--names]")
	}

	return g.DeleteImages(df)
}

// Modify renames the given images
func (g *GceCommand) Modify(args []string) error {
	m := newDeprecateOptions()
	if err := m.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		m.flagSet.Usage()
		return nil
	}

	if len(m.Names) == 0 {
		return errors.New("no images are passed with [--names]")
	}

	return g.DeprecateImages(m)
}

// Help prints the help message for the given command
func (g *GceCommand) Help(command string) string {
	var help string
	switch command {
	case "delete":
		help = newDeleteOptions().helpMsg
	case "modify":
		help = newDeprecateOptions().helpMsg
	case "list":
		help = newListFlags().helpMsg
	default:
		return "no help found for command " + command
	}

	global := `
  -project-id      "..."              Project Id (env: IMAGES_GCE_PROJECT_ID)
  -account-file    "..."              Account file (env: IMAGES_GCE_ACCOUNT_FILE)
`

	help += global
	return help
}
