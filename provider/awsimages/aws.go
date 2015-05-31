package awsimages

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/aws/credentials"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/fatih/images/command/loader"
	"github.com/hashicorp/go-multierror"
	"github.com/shiena/ansicolor"
)

type awsConfig struct {
	// just so we can use the Env and TOML loader more efficiently with out
	// any complex hacks
	Aws struct {
		Region        string `toml:"region" json:"region"`
		RegionExclude string `toml:"region_exclude" json:"region_exclude"`
		AccessKey     string `toml:"access_key" json:"access_key"`
		SecretKey     string `toml:"secret_key" json:"secret_key"`
	}
}

// AwsImages is responsible of managing AWS images (AMI's)
type AwsImages struct {
	services *multiRegion
	images   map[string][]*ec2.Image
}

// New returns a new instance of AwsImages
func New(args []string) (*AwsImages, error) {
	conf := new(awsConfig)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	checkCfg := "Please check your configuration"

	if conf.Aws.Region == "" {
		return nil, errors.New("AWS Region is not set. " + checkCfg)
	}

	if conf.Aws.AccessKey == "" {
		return nil, errors.New("AWS Access Key is not set. " + checkCfg)
	}

	if conf.Aws.SecretKey == "" {
		return nil, errors.New("AWS Secret Key is not set. " + checkCfg)
	}

	// increase the timeout
	timeout := time.Second * 30
	client := &http.Client{
		Transport: &http.Transport{TLSHandshakeTimeout: timeout},
		Timeout:   timeout,
	}

	creds := credentials.NewStaticCredentials(conf.Aws.AccessKey, conf.Aws.SecretKey, "")
	awsCfg := &aws.Config{
		Credentials: creds,
		HTTPClient:  client,
		Logger:      os.Stdout,
	}

	m := newMultiRegion(awsCfg, parseRegions(conf.Aws.Region, conf.Aws.RegionExclude))
	return &AwsImages{
		services: m,
		images:   make(map[string][]*ec2.Image),
	}, nil
}

// Fetch fetches the given images and stores them internally. Call Print()
// method to output them.
func (a *AwsImages) Fetch(args []string) error {
	input := &ec2.DescribeImagesInput{
		Owners: stringSlice("self"),
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	for r, s := range a.services.regions {
		wg.Add(1)
		go func(region string, svc *ec2.EC2) {
			resp, err := svc.DescribeImages(input)
			mu.Lock()

			if err != nil {
				multiErrors = multierror.Append(multiErrors, err)
			} else {
				// sort from oldest to newest
				if len(resp.Images) > 1 {
					sort.Sort(byTime(resp.Images))
				}

				a.images[region] = resp.Images
			}

			mu.Unlock()
			wg.Done()
		}(r, s)
	}

	wg.Wait()

	return multiErrors
}

// Print prints the stored images to standard output.
func (a *AwsImages) Print() {
	if len(a.images) == 0 {
		fmt.Fprintln(os.Stderr, "no images found")
		return
	}

	green := color.New(color.FgGreen).SprintfFunc()

	w := new(tabwriter.Writer)
	w.Init(ansicolor.NewAnsiColorWriter(os.Stdout), 10, 8, 0, '\t', 0)
	defer w.Flush()

	for region, images := range a.images {
		if len(images) == 0 {
			continue
		}

		fmt.Fprintln(w, green("AWS Region: %s (%d images):", region, len(images)))
		fmt.Fprintln(w, "    Name\tID\tState\tTags")

		for i, image := range images {
			tags := make([]string, len(image.Tags))
			for i, tag := range image.Tags {
				tags[i] = *tag.Key + ":" + *tag.Value
			}

			name := ""
			if image.Name != nil {
				name = *image.Name
			}

			state := *image.State
			if *image.State == "failed" {
				state += " (" + *image.StateReason.Message + ")"
			}

			fmt.Fprintf(w, "[%d] %s\t%s\t%s\t%+v\n",
				i+1, name, *image.ImageID, state, tags)
		}

		fmt.Fprintln(w, "")
	}
}

// Help prints the help message for the given command
func (a *AwsImages) Help(command string) string {
	var help string

	global := `
  -access-key      "..."       AWS Access Key (env: IMAGES_AWS_ACCESS_KEY)
  -secret-key      "..."       AWS Secret Key (env: IMAGES_AWS_SECRET_KEY)
  -region          "..."       AWS Region (env: IMAGES_AWS_REGION)
  -region-exclude  "..."       AWS Region to be excluded (env: IMAGES_AWS_REGION_EXCLUDE)
`
	switch command {
	case "modify":
		help = newModifyFlags().helpMsg
	case "delete":
		help = newDeleteFlags().helpMsg
	case "list":
		help = `Usage: images list --provider aws [options]

 List AMI properties.

Options:
	`
	case "copy":
		help = newCopyFlags().helpMsg
	default:
		return "no help found for command " + command
	}

	help += global
	return help
}
