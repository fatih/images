package sl

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"provider/utils"

	"github.com/fatih/flags"
)

type listFlags struct {
	output   utils.OutputMode
	imageIds []int
	all      bool

	helpMsg string
	flagSet *flag.FlagSet
}

func newListFlags() *listFlags {
	l := &listFlags{}

	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	flagSet.BoolVar(&l.all, "all", false, "Display system and not taggable images.")
	flagSet.Var(utils.NewOutputValue(utils.Simplified, &l.output), "output", "Output mode")
	flagSet.Var(flags.NewIntSlice(nil, &l.imageIds), "ids", "Images to be listed. Default case is all images")
	l.helpMsg = `Usage: images list --providers sl [options]

   List AMI properties.

Options:

  -ids     "123,..."   Images to be listed. By default all images are shown.
  -all                 Display all images - systems ones (-SWAP, -METADATA)
                       and not taggable ones as well.
                       By default only taggable images are displayed.
  -output  "json"      Output mode of images. (default: "simplified")
                       Available options: "json" or "simplified"
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, l.helpMsg)
	}
	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	l.flagSet = flagSet
	return l
}

// hack for broken softlayer-go API
var empty = &bytes.Buffer{}

// Images returns all images. If not images are found, it returns non-nil error.
func (img *SLImages) Images() (Images, error) {
	var images []*Image
	path := fmt.Sprintf("%s/getBlockDeviceTemplateGroups.json", img.account.GetName())
	p, err := img.client.DoRawHttpRequestWithObjectMask(path, imageMask, "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(p, &images); err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, errors.New("no images found")
	}

	sort.Sort(Images(images))
	for _, image := range images {
		image.decode()
	}

	return images, nil
}

// ImagesByIDs looks up all images and then it filters them by the given
// IDs. If at least one image is not found, it returns non-nil error.
func (img *SLImages) ImagesByIDs(ids ...int) (Images, error) {
	images, err := img.Images()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return images, nil
	}

	filter := make(map[int]struct{}, len(ids))
	filtered := []*Image{}
	for _, id := range ids {
		filter[id] = struct{}{}
	}
	for _, img := range images {
		if _, ok := filter[img.ID]; ok {
			filtered = append(filtered, img)
			delete(filter, img.ID)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no images found for ids=%v", ids)
	}

	if len(filter) != 0 {
		ids = make([]int, 0, len(filter))
		for id := range filter {
			ids = append(ids, id)
		}
		return nil, fmt.Errorf("the following images were not found: %v", ids)
	}

	return filtered, nil
}

// ImageByID looks up an image by the given ID.
func (img *SLImages) ImageByID(id int) (*Image, error) {
	path := fmt.Sprintf("%s/%d/getObject.json", img.block.GetName(), id)
	p, err := img.client.DoRawHttpRequestWithObjectMask(path, imageMask, "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	var image Image
	if err = json.Unmarshal(p, &image); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response: %s", err)
	}

	image.decode()
	return &image, nil
}
