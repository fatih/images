package sl

import (
	"command/loader"
	"errors"
	"strings"
)

// SLCommand implements the images various interfaces, such as Fetcher,
// Deleter, Modifier, etc..
type SLCommand struct {
	*SLImages
}

// NewCommand returns a new instance of SLCommand
func NewCommand(args []string) (*SLCommand, []string, error) {
	var conf struct {
		// just so we can use the Env and TOML loader more efficiently with out
		// any complex hacks
		SL SLConfig
	}

	if err := loader.Load(&conf, args); err != nil {
		return nil, nil, err
	}

	slImages, err := New(&conf.SL)
	if err != nil {
		return nil, nil, err
	}

	remainingArgs := loader.ExcludeArgs(&conf, args)

	return &SLCommand{
		SLImages: slImages,
	}, remainingArgs, nil
}

// List implements the command.Lister interface
func (cmd *SLCommand) List(args []string) error {
	l := newListFlags()
	if err := l.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(l.imageIds) == 1 {
		image, err := cmd.ImageByID(l.imageIds[0])
		if err != nil {
			return err
		}
		return (Images{image}).Print(l.output)
	}

	if len(l.imageIds) != 0 {
		images, err := cmd.ImagesByIDs(l.imageIds...)
		if err != nil {
			return err
		}
		return images.Print(l.output)
	}

	images, err := cmd.Images()
	if err != nil {
		return err
	}

	if l.all {
		return images.Print(l.output)
	}

	var filtered Images
	// Filter out system images and not taggable ones.
	for _, img := range images {
		isSystem := strings.HasSuffix(img.Name, "-SWAP") || strings.HasSuffix(img.Name, "-METADATA")
		if isSystem || img.NotTaggable {
			continue
		}
		filtered = append(filtered, img)
	}
	return filtered.Print(l.output)
}

// Modify manages the tags of the given images. It can create, override or
// delete tags associated with the given Template ids.
func (cmd *SLCommand) Modify(args []string) error {
	l := newModifyFlags()
	if err := l.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(l.imageIds) == 0 {
		return errors.New("no value for -ids flag")
	}

	createTags := newTags(l.createTags)
	deleteTags := newTags(l.deleteTags)
	if len(createTags) != 0 && len(deleteTags) != 0 {
		patchFn := func(orig Tags) {
			for k, v := range createTags {
				orig[k] = v
			}
			for k := range deleteTags {
				delete(orig, k)
			}
		}
		return cmd.patchTags(patchFn, l.force, l.imageIds...)
	} else if len(createTags) != 0 {
		return cmd.createTags(createTags, l.force, l.imageIds...)
	} else if len(deleteTags) != 0 {
		return cmd.deleteTags(deleteTags, l.force, l.imageIds...)
	}
	return errors.New("neither -create-tags nor -delete-tags flag was specified")
}

// Delete deletes Block Device Templates by the given ids.
func (cmd *SLCommand) Delete(args []string) error {
	l := newModifyFlags()
	if err := l.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(l.imageIds) == 0 {
		return errors.New("no value for -ids flag")
	}

	return cmd.DeleteImages(l.imageIds...)
}

// Help prints the help message for the given command
func (a *SLCommand) Help(command string) string {
	var help string

	global := `
  -username        "..."       Sofleyer Username (env: IMAGES_SL_USERNAME)
  -api-key         "..."       Softlayer API Key (env: IMAGES_SL_API_KEY)
`
	switch command {
	case "modify":
		help = newModifyFlags().helpMsg
	case "list":
		help = newListFlags().helpMsg
	case "delete":
		help = newDeleteFlags().helpMsg
	default:
		return "no help found for command " + command
	}

	help += global
	return help
}
