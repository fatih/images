package sl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"provider/utils"

	"github.com/fatih/color"
	"github.com/shiena/ansicolor"
)

// Datacenter represents a Softlayer datacenter.
type Datacenter struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	StatusID int    `json:"statusId,omitempty"`
}

// Image
type Image struct {
	ID          int           `json:"id,omitempty"`
	ParentID    int           `json:"parentId,omitempty"`
	GlobalID    string        `json:"globalIdentifier,omitempty"`
	CreateDate  time.Time     `json:"createDate,omitempty"`
	Name        string        `json:"name,omitempty"`
	Note        string        `json:"note,omitempty"`
	Datacenter  *Datacenter   `json:"datacenter,omitempty"`
	Datacenters []*Datacenter `json:"datacenters,omitempty"`

	Tags        Tags `json:"-"`
	NotTaggable bool `json:"-"`
}

func (img *Image) globalID() string {
	if img.GlobalID != "" {
		return img.GlobalID
	}
	return "-"
}

func (img *Image) datacenters() string {
	var names []string
	if img.Datacenter != nil {
		names = append(names, img.Datacenter.Name)
	}
	for _, d := range img.Datacenters {
		names = append(names, d.Name)
	}
	if len(names) == 0 {
		return "-"
	}
	return strings.Join(names, ",")
}

func (img *Image) tags() string {
	if img.NotTaggable {
		return "-"
	}
	return img.Tags.String()
}

// decode unmarshals tags from description or mark as non taggable when decoding fails.
func (img *Image) decode() {
	if err := json.Unmarshal([]byte(img.Note), &img.Tags); err != nil {
		img.NotTaggable = true
	}
}

// encode marshals tags from description field.
func (img *Image) encode() error {
	if !img.NotTaggable && len(img.Tags) != 0 {
		p, err := json.Marshal((map[string]string)(img.Tags))
		if err != nil {
			return fmt.Errorf("unable to marshal tags: %s", err)
		}
		img.Note = string(p)
	}
	return nil
}

var imageMask = []string{
	"id",
	"parentId",
	"globalIdentifier",
	"createDate",
	"name",
	"note",
	"datacenter",
	"datacenters",
}

// Images defines and represents regions to images
type Images []*Image

// Len, Less and Swap implements the sort.Interface interface.
func (img Images) Len() int           { return len(img) }
func (img Images) Less(i, j int) bool { return img[i].ID < img[i].ID }
func (img Images) Swap(i, j int)      { img[i], img[j] = img[j], img[i] }

// Print prints the images to standard output.
func (img Images) Print(mode utils.OutputMode) error {
	if len(img) == 0 {
		return errors.New("no images found (use -all flag)")
	}

	switch mode {
	case utils.JSON:
		p, err := json.MarshalIndent(img, "", "    ")
		if err != nil {
			return err
		}

		fmt.Println(string(p))
		return nil
	case utils.Simplified:
		green := color.New(color.FgGreen).SprintfFunc()
		w := utils.NewImagesTabWriter(ansicolor.NewAnsiColorWriter(os.Stdout))
		defer w.Flush()

		fmt.Fprintln(w, green("Softlayer (%d images):", len(img)))
		fmt.Fprintln(w, "    Name\tID\tGlobalID\tCreated\tDatacenters\tTags")

		for i, image := range img {
			created := "-"
			if !image.CreateDate.IsZero() {
				created = image.CreateDate.Format(time.RFC3339)
			}
			fmt.Fprintf(w, "[%d] %s\t%d\t%s\t%s\t%s\t%s\n", i, image.Name, image.ID,
				image.globalID(), created, image.datacenters(), image.tags())
		}

		fmt.Fprintln(w)
		return nil
	default:
		return fmt.Errorf("output mode %q is not valid", mode)
	}
	return nil
}
