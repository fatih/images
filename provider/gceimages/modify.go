package gceimages

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	compute "google.golang.org/api/compute/v1"

	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type DeprecateOptions struct {
	Names []string
	State string

	helpMsg string
	flagSet *flag.FlagSet
}

func newDeprecateOptions() *DeprecateOptions {
	m := &DeprecateOptions{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.Var(flags.NewStringSlice(nil, &m.Names), "ids", "Images to be delete with the given names")
	flagSet.StringVar(&m.State, "state", "", "Image state to be applied")
	m.helpMsg = `Usage: images modify --provider gce [options]

  Depcreate images

Options:

  -names    "myImage,..."   Images to be deprecated
  -state    "..."           Image state to be applied. Possible values:
                            DELETED, DEPRECATED, OBSOLETE or "" (to clear state)
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, m.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	m.flagSet = flagSet
	return m
}

// Modify renames the given images
func (g *GceImages) DeprecateImages(opts *DeprecateOptions) error {
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for _, n := range opts.Names {
		wg.Add(1)
		go func(name string) {
			st := &compute.DeprecationStatus{
				State: opts.State,
			}

			_, err := g.svc.Deprecate(g.config.ProjectID, name, st).Do()
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
