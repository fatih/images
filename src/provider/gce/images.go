package gce

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"provider/utils"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
	compute "google.golang.org/api/compute/v1"
)

type Images compute.ImageList

// Print prints the stored images to standard output.
func (i Images) Print(mode utils.OutputMode) error {
	if len(i.Items) == 0 {
		return errors.New("no images found")
	}

	switch mode {
	case utils.JSON:
		out, err := i.outputJSON()
		if err != nil {
			return err
		}

		fmt.Println(out)
		return nil
	case utils.Simplified:
		green := color.New(color.FgGreen).SprintfFunc()
		output := ansicolor.NewAnsiColorWriter(os.Stdout)
		w := utils.NewImagesTabWriter(output)
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
	default:
		return fmt.Errorf("output mode '%s' is not valid", mode)
	}
}

// outputJSON returns a JSON formatted output of all images
func (i Images) outputJSON() (string, error) {
	out, err := json.MarshalIndent(&i, "", "    ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}
