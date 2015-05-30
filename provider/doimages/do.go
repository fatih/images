package doimages

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/digitalocean/godo"
	"github.com/fatih/images/command/loader"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type tokenSource struct {
	AccessToken string
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

type DoConfig struct {
	// just so we can use the Env and TOML loader more efficiently with out
	// any complex hacks
	Do struct {
		Token string `toml:"token" json:"token"`
	}
}

type DoImages struct {
	client *godo.Client
	images map[string][]godo.Image
}

func New(args []string) *DoImages {
	conf := new(DoConfig)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	if conf.Do.Token == "" {
		fmt.Fprintln(os.Stderr, "token is not set")
		os.Exit(1)
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
		AccessToken: conf.Do.Token,
	})

	godoClient := godo.NewClient(oauthClient)

	return &DoImages{
		client: godoClient,
		images: make(map[string][]godo.Image),
	}
}
