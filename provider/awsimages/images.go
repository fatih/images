package awsimages

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

// Images defines and represents regions to images
type Images map[string][]*ec2.Image

// Print prints the images to standard output.
func (i Images) Print() error {
	if len(i) == 0 {
		return errors.New("no images found")
	}

	green := color.New(color.FgGreen).SprintfFunc()
	output := ansicolor.NewAnsiColorWriter(os.Stdout)

	w := tabwriter.NewWriter(output, 10, 8, 0, '\t', 0)
	defer w.Flush()

	for region, images := range i {
		if len(images) == 0 {
			continue
		}

		fmt.Fprintln(w, green("AWS Region: %s (%d images):", region, len(images)))
		fmt.Fprintln(w, "    Name\tID\tState\tTags")

		for ix, image := range images {
			tags := make([]string, len(image.Tags))
			for i, tag := range image.Tags {
				tags[i] = *tag.Key + ":" + *tag.Value
			}

			name := ""
			if image.Name != nil {
				name = *image.Name
			}

			state := *image.State
			if *image.State == "failed" {
				state += " (" + *image.StateReason.Message + ")"
			}

			fmt.Fprintf(w, "[%d] %s\t%s\t%s\t%+v\n",
				ix+1, name, *image.ImageID, state, tags)
		}

		fmt.Fprintln(w, "")
	}

	return nil
}

// RegionFromId returns the region for the given id
func (i Images) RegionFromId(id string) (string, error) {
	if len(i) == 0 {
		return "", errors.New("images are not fetched")
	}

	for region, images := range i {
		for _, image := range images {
			if *image.ImageID == id {
				return region, nil
			}
		}
	}

	return "", fmt.Errorf("no region found for image id '%s'", id)
}
