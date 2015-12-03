package sl

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

// Tags holds key-value tags for an image.
type Tags map[string]string

// String gives key-value tags representation.
func (t Tags) String() string {
	if len(t) == 0 {
		return ""
	}
	var buf bytes.Buffer
	fmt.Fprint(&buf, "[")
	for k, v := range t {
		fmt.Fprint(&buf, k, "=", v, ",")
	}
	p := buf.Bytes()
	p[len(p)-1] = ']' // replace last dangling commna
	return string(p)
}

func newTags(kv []string) Tags {
	if len(kv) == 0 {
		return nil
	}
	t := make(Tags)
	for _, kv := range kv {
		if i := strings.IndexRune(kv, '='); i != -1 {
			t[kv[:i]] = kv[i+1:]
		} else {
			t[kv] = ""
		}
	}
	return t
}

type modifyFlags struct {
	createTags []string
	deleteTags []string
	imageIds   []int
	force      bool
	helpMsg    string

	flagSet *flag.FlagSet
}

func newModifyFlags() *modifyFlags {
	m := &modifyFlags{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.BoolVar(&m.force, "f", false, "Force creation of tags on not taggable image")
	flagSet.Var(flags.NewStringSlice(nil, &m.createTags), "create-tags", "Create  or override tags")
	flagSet.Var(flags.NewStringSlice(nil, &m.deleteTags), "delete-tags", "Delete tags")
	flagSet.Var(flags.NewIntSlice(nil, &m.imageIds), "ids", "Images to be delete with actions")
	m.helpMsg = `Usage: images modify --providers sl [options]

  Modify AMI properties.

Options:

  -ids         "123,..."   Images to be used with below actions
  -create-tags "key=val,..."   Create or override tags
  -delete-tags "key,..."       Delete tags
  -f                           Force creation of tags on not taggable image.
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, m.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	m.flagSet = flagSet
	return m
}

func (img *SLImages) createTags(tags Tags, force bool, imageIDs ...int) error {
	if len(tags) == 0 {
		return errors.New("not tags to create")
	}
	patchFn := func(orig Tags) {
		for k, v := range tags {
			orig[k] = v
		}
	}
	return img.patchTags(patchFn, force, imageIDs...)
}

func (img *SLImages) deleteTags(tags Tags, force bool, imageIDs ...int) error {
	if len(tags) == 0 {
		return errors.New("not tags to delete")
	}
	patchFn := func(orig Tags) {
		for k := range tags {
			delete(orig, k)
		}
	}
	return img.patchTags(patchFn, force, imageIDs...)
}

func (img *SLImages) patchTags(patchFn func(orig Tags), force bool, imageIDs ...int) error {
	images, err := img.ImagesByIDs(imageIDs...)
	if err != nil {
		return err
	}
	for _, image := range images {
		if image.NotTaggable && !force {
			err = multierror.Append(err, fmt.Errorf("unable to patch not taggable image with id=%d (use -f to override)", image.ID))
			continue
		}
		fields := &Image{
			Tags: image.Tags,
		}
		if fields.Tags == nil {
			fields.Tags = make(Tags)
		}
		patchFn(fields.Tags)
		if e := img.EditImage(image.ID, fields); e != nil {
			err = multierror.Append(err, fmt.Errorf("failed to patch image with id=%d: %s", image.ID, e))
		}
	}
	return err

}
