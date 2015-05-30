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

	"github.com/hashicorp/go-multierror"
)

type deleteFlags struct {
	imageIds string
	helpMsg  string

	flagSet *flag.FlagSet
}

func newDeleteFlags() *deleteFlags {
	d := &deleteFlags{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.StringVar(&d.imageIds, "image-ids", "", "Images to be deleted with the given ids")
	d.helpMsg = `Usage: images delete --provider do [options]

  Delete images

Options:

  -image-ids   "123,..."   Images to be deleted with the given ids
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, d.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	d.flagSet = flagSet
	return d
}

func (d *DoImages) Delete(args []string) error {
	df := newDeleteFlags()

	if err := df.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		df.flagSet.Usage()
		return nil
	}

	if df.imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	images := strings.Split(df.imageIds, ",")

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

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
			_, err := d.client.Images.Delete(id)
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
