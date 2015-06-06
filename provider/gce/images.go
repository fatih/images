package gce

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	compute "google.golang.org/api/compute/v1"
)

type Images compute.ImageList

// Print prints the stored images to standard output.
func (i Images) Print() error {
	if len(i.Items) == 0 {
		return errors.New("no images found")
	}

	green := color.New(color.FgGreen).SprintfFunc()
	output := ansicolor.NewAnsiColorWriter(os.Stdout)
	w := tabwriter.NewWriter(output, 10, 8, 0, '\t', 0)
	defer w.Flush()

	imageDesc := "image"
	if len(i.Items) > 1 {
		imageDesc = "images"
	}

	fmt.Fprintln(w, green("GCE (%d %s):", len(i.Items), imageDesc))
	fmt.Fprintln(w, "    Name\tID\tStatus\tType\tDeprecated\tCreation Timestamp")

	for ix, image := range i.Items {
		deprecatedState := ""
		if image.Deprecated != nil {
			deprecatedState = image.Deprecated.State
		}

		fmt.Fprintf(w, "[%d] %s (%s)\t%d\t%s\t%s (%d)\t%s\t%s\n",
			ix+1, image.Name, image.Description, image.Id,
			image.Status, image.SourceType, image.DiskSizeGb,
			deprecatedState, image.CreationTimestamp,
		)
	}

	return nil
}
