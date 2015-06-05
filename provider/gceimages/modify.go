package gceimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	compute "google.golang.org/api/compute/v1"

	"github.com/hashicorp/go-multierror"
)

type modifyFlags struct {
	names   string
	state   string
	helpMsg string

	flagSet *flag.FlagSet
}

func newModifyFlags() *modifyFlags {
	m := &modifyFlags{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&m.names, "names", "", "Images to be deprecated")
	flagSet.StringVar(&m.state, "state", "", "Image state to be applied")
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
func (g *GceImages) Modify(args []string) error {
	m := newModifyFlags()
	if err := m.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		m.flagSet.Usage()
		return nil
	}

	if m.names == "" {
		return errors.New("no images are passed with [--names]")
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	images := strings.Split(m.names, ",")

	for _, n := range images {
		wg.Add(1)
		go func(name string) {
			st := &compute.DeprecationStatus{
				State: m.state,
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
