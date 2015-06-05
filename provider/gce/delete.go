package gce

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
	Names []string

	helpMsg string
	flagSet *flag.FlagSet
}

func newDeleteOptions() *DeleteOptions {
	d := &DeleteOptions{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.Var(flags.NewStringSlice(nil, &d.Names), "names", "Images to be delete with the given names")
	d.helpMsg = `Usage: images delete --provider gce [options]

  Delete images

Options:

  -names           "myImage,..."      Images to be deleted with the given names
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, d.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	d.flagSet = flagSet
	return d
}

// Delete deletes the given images.
func (g *GceImages) DeleteImages(opts *DeleteOptions) error {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for _, n := range opts.Names {
		wg.Add(1)
		go func(name string) {
			_, err := g.svc.Delete(g.config.ProjectID, name).Do()
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}

			wg.Done()
		}(n)
	}

	wg.Wait()
	return multiErrors
}
