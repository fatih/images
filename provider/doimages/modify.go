package doimages

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/digitalocean/godo"
	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type RenameOptions struct {
	// ImageIds to be renamed
	ImageIds []int

	// Name the images will changed to
	Name string

	helpMsg string
	flagSet *flag.FlagSet
}

func newRenameOptions() *RenameOptions {
	r := &RenameOptions{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.Var(flags.NewIntSlice(nil, &r.ImageIds), "ids", "Images to be delete with the given ids")
	flagSet.StringVar(&r.Name, "name", "", "New name for the images")
	r.helpMsg = `Usage: images modify --provider do [options]

  Rename images

Options:

  -ids         "123,..."       Images to be renamed
  -name        "example"       New name for the images
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, r.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	r.flagSet = flagSet
	return r
}

// RenameImages renames the images to the given new names
func (d *DoImages) RenameImages(opts *RenameOptions) error {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for _, imageID := range opts.ImageIds {
		wg.Add(1)
		go func(id int) {
			_, _, err := d.client.Images.Update(id, &godo.ImageUpdateRequest{Name: opts.Name})
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}

			wg.Done()
		}(imageID)
	}

	wg.Wait()
	return multiErrors
}
