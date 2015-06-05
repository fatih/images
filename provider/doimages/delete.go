package doimages

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type DeleteOptions struct {
	ImageIds []int
	helpMsg  string

	flagSet *flag.FlagSet
}

func newDeleteOptions() *DeleteOptions {
	d := &DeleteOptions{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.Var(flags.NewIntSlice(nil, &d.ImageIds), "ids", "Images to be delete with the given ids")
	d.helpMsg = `Usage: images delete --provider do [options]

  Delete images

Options:

  -ids         "123,..."       Images to be deleted with the given ids
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, d.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	d.flagSet = flagSet
	return d
}

// Delete deletes the given images.
func (d *DoImages) DeleteImages(opts *DeleteOptions) error {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for _, imageID := range opts.ImageIds {
		wg.Add(1)
		go func(id int) {
			_, err := d.client.Images.Delete(id)
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
