package sl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

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

// Transaction returns an ongoing transaction for the image given by the id.
//
// It returns non-nil error when querying the service failed.
// It return nil Transaction and nil error when there are no ongoing
// transactions.
func (img *SLImages) Transaction(id int) (*Transaction, error) {
	path := fmt.Sprintf("%s/%d/getTransaction.json", img.block.GetName(), id)
	p, err := img.client.DoRawHttpRequest(path, "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	if bytes.Compare(p, []byte("null")) == 0 {
		return nil, nil
	}

	var t Transaction
	if err = json.Unmarshal(p, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

// WaitReady waits at most d until all ongoing transactions for image
// given by the id are finished.
func (img *SLImages) WaitReady(id int, d time.Duration) error {
	timeout := time.After(d)
	parent, err := img.Parent(id)
	if err != nil {
		return err
	}

	if parent != nil && parent.ID != 0 {
		id = parent.ID
	}

	children, err := img.Children(id)
	if err != nil {
		return err
	}

	ids := make([]int, len(children))
	for i, child := range children {
		ids[i] = child.ID
	}
	ids = append(ids, id)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("waiting for %d to be ready timed out after %s", id, d)
		default:
			var ongoing bool

			for _, id := range ids {
				t, err := img.Transaction(id)
				if err != nil {
					return err
				}

				if t != nil && t.ID != 0 {
					ongoing = true
					break
				}
			}

			if !ongoing {
				return nil
			}

			time.Sleep(5 * time.Second)
		}
	}
}

// Parent returns a parent image for the one given by the id.
//
// It returns nil Image and nil error if the image has no parent.
func (img *SLImages) Parent(id int) (*Image, error) {
	path := fmt.Sprintf("%s/%d/getParent.json", img.block.GetName(), id)
	p, err := img.client.DoRawHttpRequest(path, "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	if bytes.Compare(p, []byte("null")) == 0 {
		return nil, nil
	}

	var image Image
	if err = json.Unmarshal(p, &image); err != nil {
		return nil, err
	}

	return &image, nil

}

// Children returns children images for the one given by the id.
//
// It returns nil Images and nil error if the images has no children.
func (img *SLImages) Children(id int) (Images, error) {
	path := fmt.Sprintf("%s/%d/getBlockDevices.json", img.block.GetName(), id)
	p, err := img.client.DoRawHttpRequest(path, "GET", empty)
	if err != nil {
		return nil, err
	}

	if err = newError(p); err != nil {
		return nil, err
	}

	var images Images
	if err = json.Unmarshal(p, &images); err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, nil
	}

	return images, nil
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
