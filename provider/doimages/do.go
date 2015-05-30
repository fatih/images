package doimages

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/digitalocean/godo"
	"github.com/fatih/color"
	"github.com/fatih/images/command/loader"
	"github.com/shiena/ansicolor"
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
	images []godo.Image
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
		images: make([]godo.Image, 0),
	}
}

func (d *DoImages) Fetch(args []string) error {
	var err error
	d.images, _, err = d.client.Images.ListUser(nil)
	return err
}

func (d *DoImages) Print() {
	if len(d.images) == 0 {
		fmt.Fprintln(os.Stderr, "no images found")
		return
	}

	green := color.New(color.FgGreen).SprintfFunc()

	w := new(tabwriter.Writer)
	w.Init(ansicolor.NewAnsiColorWriter(os.Stdout), 10, 8, 0, '\t', 0)
	defer w.Flush()

	imageDesc := "image"
	if len(d.images) > 1 {
		imageDesc = "images"
	}

	fmt.Fprintln(w, green("DO: (%d %s):", len(d.images), imageDesc))

	for i, image := range d.images {
		fmt.Fprintln(w, "    Name\tID\tDistribution\tType\tRegions")

		regions := make([]string, len(image.Regions))
		for i, region := range image.Regions {
			regions[i] = region
		}

		fmt.Fprintf(w, "[%d] %s\t%d\t%s\t%s (%d)\t%+v\n",
			i, image.Name, image.ID, image.Distribution, image.Type, image.MinDiskSize, regions)

		fmt.Fprintln(w, "")
	}
}
