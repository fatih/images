package doimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/go-multierror"
)

type copyFlags struct {
	imageID       string
	sourceRegions string
	desc          string
	dryRun        bool
	helpMsg       string

	flagSet *flag.FlagSet
}

func newCopyFlags() *copyFlags {
	c := &copyFlags{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.StringVar(&c.imageID, "image", "", "Image to be copied with the given id")
	flagSet.StringVar(&c.sourceRegions, "to", "", "Images to be copied to the given regions")

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

func (d *DoImages) Copy(args []string) error {
	c := newCopyFlags()

	if err := c.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		c.flagSet.Usage()
		return nil
	}

	if c.imageID == "" {
		return errors.New("no image is passed. Use --image")
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	imageId, err := strconv.Atoi(c.imageID)
	if err != nil {
		return err
	}

	regions := strings.Split(c.sourceRegions, ",")

	for _, r := range regions {
		wg.Add(1)
		go func(region string) {
			_, _, err := d.client.ImageActions.Transfer(imageId, &godo.ActionRequest{
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
