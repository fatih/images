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

type CopyOptions struct {
	ImageID       int
	SourceRegions []string

	helpMsg string
	flagSet *flag.FlagSet
}

func newCopyOptions() *CopyOptions {
	c := &CopyOptions{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.IntVar(&c.ImageID, "image", 0, "Image to be copied with the given id")
	flagSet.Var(flags.StringListVar(&c.SourceRegions), "to", "Images to be copied to the given regions")

	c.helpMsg = `Usage: images copy --provider do [options]

  Copy image to regions

Options:

  -image   "123"               Image to be copied with the given id
  -to      "fra1,nyc2,..."     Image to be copied to the given regions 
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, c.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	c.flagSet = flagSet
	return c
}

// Copy transfers the images to other regions
func (d *DoImages) CopyImages(opts *CopyOptions) error {
	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	for _, r := range opts.SourceRegions {
		wg.Add(1)
		go func(region string) {
			_, _, err := d.client.ImageActions.Transfer(opts.ImageID, &godo.ActionRequest{
				"type":   "transfer",
				"region": region,
			})
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}

			wg.Done()
		}(r)
	}

	wg.Wait()
	return multiErrors
}
