package doimages

import (
	"errors"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type DoConfig struct {
	Token string `toml:"token" json:"token"`
}

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// DoImages is responsible of managing DigitalOcean images
type DoImages struct {
	client *godo.Client
	images []godo.Image
}

// New returns a new instance of DoImages
func New(conf *DoConfig) (*DoImages, error) {
	if conf.Token == "" {
		return nil, errors.New("Access Token is not set. Please check your configuration.")
	}

	// increase the timeout
	timeout := time.Second * 30
	client := &http.Client{
		Transport: &http.Transport{TLSHandshakeTimeout: timeout},
		Timeout:   timeout,
	}

	// we need to pass the client with the context itself
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, client)

	oauthClient := oauth2.NewClient(ctx, &tokenSource{
		AccessToken: conf.Token,
	})

	godoClient := godo.NewClient(oauthClient)

	return &DoImages{
		client: godoClient,
		images: make([]godo.Image, 0),
	}, nil
}

func (d *DoImages) UserImages() (Images, error) {
	images, _, err := d.client.Images.ListUser(nil)
	return Images(images), err
}
