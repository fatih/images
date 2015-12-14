package sl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	slclient "github.com/maximilien/softlayer-go/client"
	"github.com/maximilien/softlayer-go/softlayer"
)

// SLConfig represents a configuration section of .imagesrc for sl provider.
type SLConfig struct {
	Username string `toml:"username" json:"username"`
	APIKey   string `toml:"api_key" json:"api_key"`
}

// SLImages is responsible of managing Softlayer Virtual Disk Images.
type SLImages struct {
	client softlayer.Client

	account softlayer.SoftLayer_Account_Service
	block   softlayer.SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service
}

// New creates new Softlayer command client.
func New(conf *SLConfig) (*SLImages, error) {
	// To suppress softlayer-go debug printfs...
	if err := os.Setenv("SL_GO_NON_VERBOSE", "YES"); err != nil {
		return nil, err
	}
	client := slclient.NewSoftLayerClient(conf.Username, conf.APIKey)
	account, err := client.GetSoftLayer_Account_Service()
	if err != nil {
		return nil, err
	}

	block, err := client.GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service()
	if err != nil {
		return nil, err
	}

	return &SLImages{
		client:  client,
		account: account,
		block:   block,
	}, nil
}

// EditImage edits non-zero fields of the image given by the id.
func (img *SLImages) EditImage(id int, fields *Image) error {
	if fields == nil {
		return nil
	}

	if err := fields.encode(); err != nil {
		return err
	}

	req := struct {
		Parameters []*Image `json:"parameters"`
	}{[]*Image{fields}}
	p, err := json.Marshal(req)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%d/editObject.json", img.block.GetName(), id)
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
		return fmt.Errorf("failed patching image=%d with fields=%+v", id, fields)
	}
	return nil
}
