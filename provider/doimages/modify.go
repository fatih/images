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

type modifyFlags struct {
	name     string
	imageIds string
	helpMsg  string

	flagSet *flag.FlagSet
}

func newModifyFlags() *modifyFlags {
	m := &modifyFlags{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&m.imageIds, "image-ids", "", "Images to be renamed")
	flagSet.StringVar(&m.name, "name", "", "New name for the images")
	m.helpMsg = `Usage: images modify --provider do [options]

  Rename images

Options:

  -image-ids   "ami-123,..."   Images to be renamed
  -name                        New name for the images
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, m.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	m.flagSet = flagSet
	return m
}

func (d *DoImages) Modify(args []string) error {
	m := newModifyFlags()
	if err := m.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		m.flagSet.Usage()
		return nil
	}

	if m.imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	if m.name == "" {
		return errors.New("no name is passed with [--name]")
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	images := strings.Split(m.imageIds, ",")

	for _, id := range images {
		imageId, err := strconv.Atoi(id)
		if err != nil {
			mu.Lock()
			multiErrors = multierror.Append(multiErrors, err)
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(id int) {
			_, _, err := d.client.Images.Update(id, &godo.ImageUpdateRequest{Name: m.name})
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}

			wg.Done()
		}(imageId)
	}

	wg.Wait()
	return multiErrors
}
