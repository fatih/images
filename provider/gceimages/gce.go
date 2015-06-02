package gceimages

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/fatih/images/command/loader"
	"github.com/mitchellh/go-homedir"
	"github.com/shiena/ansicolor"
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

	svc    *compute.ImagesService
	images *compute.ImageList
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
	g.images, err = g.svc.List(g.Gce.ProjectID).Do()
	return err
}

// Print prints the stored images to standard output.
func (g *GceImages) Print() {
	if len(g.images.Items) == 0 {
		fmt.Fprintln(os.Stderr, "no images found")
		return
	}

	green := color.New(color.FgGreen).SprintfFunc()

	w := new(tabwriter.Writer)
	w.Init(ansicolor.NewAnsiColorWriter(os.Stdout), 10, 8, 0, '\t', 0)
	defer w.Flush()

	imageDesc := "image"
	if len(g.images.Items) > 1 {
		imageDesc = "images"
	}

	fmt.Fprintln(w, green("GCE (%d %s):", len(g.images.Items), imageDesc))
	fmt.Fprintln(w, "    Name\tID\tStatus\tType\tDeprecated\tCreation Timestamp")

	for i, image := range g.images.Items {
		deprecatedState := ""
		if image.Deprecated != nil {
			deprecatedState = image.Deprecated.State
		}

		fmt.Fprintf(w, "[%d] %s (%s)\t%d\t%s\t%s (%d)\t%s\t%s\n",
			i+1, image.Name, image.Description, image.Id,
			image.Status, image.SourceType, image.DiskSizeGb,
			deprecatedState, image.CreationTimestamp,
		)
	}

}

// Help prints the help message for the given command
func (g *GceImages) Help(command string) string {
	var help string
	switch command {
	case "delete":
		help = newDeleteFlags().helpMsg
	case "modify":
		help = newModifyFlags().helpMsg
	case "list":
		help = `Usage: images list --provider gce [options]

 List images

Options:
	`
	default:
		return "no help found for command " + command
	}

	global := `
  -project-id      "..."              Project Id (env: IMAGES_GCE_PROJECT_ID)
  -account-file    "..."              Account file (env: IMAGES_GCE_ACCOUNT_FILE)
`

	help += global
	return help
}
