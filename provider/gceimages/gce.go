package gceimages

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fatih/images/command/loader"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
)

// GceImages is responsible of managing GCE images
type GceImages struct {
	// just so we can use the Env and TOML loader more efficiently with out
	// any complex hacks
	Gce struct {
		ProjectID   string `toml:"project_id" json:"project_id"`
		AccountFile string `toml:"account_file" json:"account_file"`
	}

	svc *compute.ImagesService
}

// New returns a new instance of GceImages
func New(args []string) (*GceImages, error) {
	cfg := new(GceImages)
	err := loader.Load(cfg, args)
	if err != nil {
		return nil, err
	}

	if cfg.Gce.ProjectID == "" {
		return nil, errors.New("ProjectID is not set. Please check your configuration.")
	}

	// increase the timeout. Also we need to pass the client with the context itself
	timeout := time.Second * 30
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, &http.Client{
		Transport: &http.Transport{TLSHandshakeTimeout: timeout},
		Timeout:   timeout,
	})

	var client *http.Client

	// allowed scopes
	scopes := []string{compute.ComputeScope}

	// Recommended way is explicit passing of credentials json which can be
	// downloaded from console.developers.google under APIs & Auth/Credentials
	// section
	if cfg.Gce.AccountFile != "" {
		// expand shell meta character
		path, err := homedir.Expand(cfg.Gce.AccountFile)
		if err != nil {
			return nil, err
		}

		jsonContent, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		jtwConfig, err := google.JWTConfigFromJSON(jsonContent, scopes...)
		if err != nil {
			return nil, err
		}

		client = jtwConfig.Client(ctx)
	} else {
		// Look for application default credentials, for more details, see:
		// https://developers.google.com/accounts/docs/application-default-credentials
		client, err = google.DefaultClient(ctx, scopes...)
		if err != nil {
			return nil, err
		}
	}

	svc, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	cfg.svc = compute.NewImagesService(svc)
	return cfg, nil
}

// Fetch fetches the given images and stores them internally. Call Print()
// method to output them.
func (g *GceImages) Fetch(args []string) error {
	var err error
	list, err := g.svc.List(g.Gce.ProjectID).Do()
	if err != nil {
		fmt.Printf("err = %+v\n", err)
		return err
	}

	fmt.Printf("list = %+v\n", list)
	for _, item := range list.Items {
		fmt.Printf("item = %+v\n", item)
	}

	return err
}

// Print prints the stored images to standard output.
func (g *GceImages) Print() {
}
