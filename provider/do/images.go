package do

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/digitalocean/godo"
	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

// Images defines and represents a list of images
type Images []godo.Image

// Print prints the stored images to standard output.
func (i Images) Print() error {
	if len(i) == 0 {
		return errors.New("no images found")
	}

	green := color.New(color.FgGreen).SprintfFunc()
	output := ansicolor.NewAnsiColorWriter(os.Stdout)
	w := tabwriter.NewWriter(output, 10, 8, 0, '\t', 0)
	defer w.Flush()

	imageDesc := "image"
	if len(i) > 1 {
		imageDesc = "images"
	}

	fmt.Fprintln(w, green("DO (%d %s):", len(i), imageDesc))
	fmt.Fprintln(w, "    Name\tID\tDistribution\tType\tRegions")

	for ix, image := range i {
		regions := make([]string, len(image.Regions))
		for i, region := range image.Regions {
			regions[i] = region
		}

		fmt.Fprintf(w, "[%d] %s\t%d\t%s\t%s (%d)\t%+v\n",
			ix+1, image.Name, image.ID, image.Distribution, image.Type, image.MinDiskSize, regions)
	}

	return nil
}
