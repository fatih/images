package gceimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type deleteFlags struct {
	names   string
	helpMsg string

	flagSet *flag.FlagSet
}

func newDeleteFlags() *deleteFlags {
	d := &deleteFlags{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.StringVar(&d.names, "names", "", "Images to be deleted with the given names")
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
func (g *GceImages) Delete(args []string) error {
	df := newDeleteFlags()

	if err := df.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		df.flagSet.Usage()
		return nil
	}

	if df.names == "" {
		return errors.New("no images are passed with [--names]")
	}

	images := strings.Split(df.names, ",")

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for _, n := range images {
		wg.Add(1)
		go func(name string) {
			_, err := g.svc.Delete(g.Gce.ProjectID, name).Do()
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
