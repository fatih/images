package sl

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/flags"
)

type copyFlags struct {
	imageID     int
	datacenters []string

	helpMsg string
	flagSet *flag.FlagSet
}

func newCopyFlags() *copyFlags {
	c := &copyFlags{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.IntVar(&c.imageID, "id", 0, "Image to be copied with the given id")
	flagSet.Var(flags.NewStringSlice(nil, &c.datacenters), "to", "Images to be copied to the given datacenters")

	c.helpMsg = `Usage: images copy --providers sl [options]

  Copy image to regions

Options:

  -id      "123"           Image to be copied with the given id
  -to      "dal05,..."     Image to be copied to the given datacenters
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, c.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	c.flagSet = flagSet
	return c
}

func (img *SLImages) datacentersByName(names ...string) ([]*Datacenter, error) {
	p, err := img.client.DoRawHttpRequest("SoftLayer_Location/getDatacenters.json", "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	var all []*Datacenter
	if err = json.Unmarshal(p, &all); err != nil {
		return nil, err
	}

	filter := make(map[string]struct{}, len(names))
	for _, name := range names {
		filter[name] = struct{}{}
	}

	filtered := make([]*Datacenter, 0, len(names))
	for _, datacenter := range all {
		if _, ok := filter[datacenter.Name]; ok {
			filtered = append(filtered, datacenter)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no datacenters found for names=%v", names)
	}

	return filtered, nil
}

func (img *SLImages) CopyToDatacenters(id int, datacenters ...string) error {
	d, err := img.datacentersByName(datacenters...)
	if err != nil {
		return err
	}
	req := struct {
		Parameters []interface{} `json:"parameters"`
	}{Parameters: []interface{}{d}}

	p, err := json.Marshal(req)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%d/addLocations.json", img.block.GetName(), id)

	p, err = img.client.DoRawHttpRequest(path, "POST", bytes.NewBuffer(p))
	if err != nil {
		return err
	}

	if err = newError(p); err != nil {
		return err
	}

	var ok bool
	if err = json.Unmarshal(p, &ok); err != nil {
		return fmt.Errorf("unable to unmarshal response: %s", err)
	}

	if !ok {
		return fmt.Errorf("failed copying image=%d to datacenters=%v", id, datacenters)
	}
	return nil
}
