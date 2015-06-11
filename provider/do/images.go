package do

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/digitalocean/godo"
	"github.com/fatih/color"
	"github.com/fatih/images/provider/utils"
	"github.com/shiena/ansicolor"
)

// Images defines and represents a list of images
type Images []godo.Image

// Print prints the stored images to standard output.
func (i Images) Print(mode utils.OutputMode) error {
	if len(i) == 0 {
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
