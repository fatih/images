package doimages

import (
	"errors"
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

func New(args []string) (*DoImages, error) {
	conf := new(DoConfig)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	if conf.Do.Token == "" {
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
		AccessToken: conf.Do.Token,
	})

	godoClient := godo.NewClient(oauthClient)

	return &DoImages{
		client: godoClient,
		images: make([]godo.Image, 0),
	}, nil
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

	fmt.Fprintln(w, green("DO (%d %s):", len(d.images), imageDesc))
	fmt.Fprintln(w, "    Name\tID\tDistribution\tType\tRegions")

	for i, image := range d.images {

		regions := make([]string, len(image.Regions))
		for i, region := range image.Regions {
			regions[i] = region
		}

		fmt.Fprintf(w, "[%d] %s\t%d\t%s\t%s (%d)\t%+v\n",
			i+1, image.Name, image.ID, image.Distribution, image.Type, image.MinDiskSize, regions)
	}
}

func (d *DoImages) Help(command string) string {
	var help string
	switch command {
	case "delete":
		help = newDeleteFlags().helpMsg
	case "modify":
		help = newModifyFlags().helpMsg
	case "copy":
		help = newCopyFlags().helpMsg
	case "list":
		help = `Usage: images list --provider do [options]

 List images

Options:
	`
	default:
		return "no help found for command " + command
	}

	global := `
  -token       "..."           DigitalOcean Access Token (env: IMAGES_DO_TOKEN)
`

	help += global
	return help
}
