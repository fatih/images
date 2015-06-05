package gce

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
)

type GCEConfig struct {
	ProjectID   string `toml:"project_id" json:"project_id"`
	AccountFile string `toml:"account_file" json:"account_file"`
}

// GceImages is responsible of managing GCE images
type GceImages struct {
	svc    *compute.ImagesService
	config *GCEConfig
}

// New returns a new instance of GceImages
func New(conf *GCEConfig) (*GceImages, error) {
	var err error
	if conf.ProjectID == "" {
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
	if conf.AccountFile != "" {
		// expand shell meta character
		path, err := homedir.Expand(conf.AccountFile)
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

	return &GceImages{
		svc:    compute.NewImagesService(svc),
		config: conf,
	}, nil
}

func (g *GceImages) ProjectImages() (Images, error) {
	images, err := g.svc.List(g.config.ProjectID).Do()
	return Images(*images), err
}
