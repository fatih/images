package sl

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type deleteFlags struct {
	imageIds []int

	helpMsg string
	flagSet *flag.FlagSet
}

func newDeleteFlags() *deleteFlags {
	d := &deleteFlags{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.Var(flags.NewIntSlice(nil, &d.imageIds), "ids", "Images to be delete with the given ids")
	d.helpMsg = `Usage: images delete --providers sl [options]

  Delete Block Device Templates.

Options:

  -ids         "123,..."   Images to be deleted with the given ids
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, d.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	d.flagSet = flagSet
	return d
}

// Delete deletes the given images.
func (img *SLImages) DeleteImages(ids ...int) error {
	var err error
	for _, id := range ids {
		path := fmt.Sprintf("%s/%d.json", img.block.GetName(), id)
		p, e := img.client.DoRawHttpRequest(path, "DELETE", empty)
		if e != nil {
			err = multierror.Append(err, fmt.Errorf("error deleting %d: %s", e))
			continue
		}

		if e := newError(p); e != nil {
			err = multierror.Append(err, fmt.Errorf("error deleting %d: %s", e))
		}
	}
	return err
}
