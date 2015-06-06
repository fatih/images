package aws

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	awsclient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type CopyOptions struct {
	// Image to be copied to other regions
	ImageID string

	// SourceRegions defines a list of regions  which the image
	// is being copied. i.e: ["us-east-1", "eu-west-1"]
	SourceRegions []string

	// Descroption for the newly created AMI (optional)
	Desc string

	// DryRun doesn't run the command, but shows the action
	DryRun bool

	helpMsg string
	flagSet *flag.FlagSet
}

func newCopyOptions() *CopyOptions {
	c := &CopyOptions{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.StringVar(&c.ImageID, "image", "", "Image to be copied with the given id")
	flagSet.StringVar(&c.Desc, "desc", "", "Description for the new AMI (optional)")
	flagSet.BoolVar(&c.DryRun, "dry-run", false, "Don't run command, but show the action")
	flagSet.Var(flags.NewStringSlice(nil, &c.SourceRegions), "to", "Images to be copied to the given regions")

	c.helpMsg = `Usage: images copy --provider aws [options]

  Copy image to regions

Options:

  -image   "ami-123"           Image to be copied with the given id
  -to      "us-east-1,..."     Image to be copied to the given regions 
  -desc    "My New Image"      Description for the new AMI's (optional)
  -dry-run                     Don't run command, but show the action
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, c.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	c.flagSet = flagSet
	return c
}

// Copy transfers the images to other regions
func (a *AwsImages) CopyImages(opts *CopyOptions) error {
	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	images, err := a.matchImages(opts.ImageID)
	if err != nil {
		return err
	}

	imageRegion := ""
	for r := range images {
		imageRegion = r
	}

	svc, err := a.svcFromRegion(imageRegion)
	if err != nil {
		return err
	}

	resp, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Owners:   stringSlice("self"),
		ImageIDs: stringSlice(opts.ImageID),
	})
	if err != nil {
		return err
	}

	image := resp.Images[0]

	if opts.Desc == "" {
		opts.Desc = *image.Description
	}

	for _, r := range opts.SourceRegions {
		wg.Add(1)
		go func(region string) {
			imageDesc := fmt.Sprintf("[Copied %s from %s via images] %s", opts.ImageID, region, opts.Desc)
			log.Println("copying image ...")
			input := &ec2.CopyImageInput{
				SourceImageID: awsclient.String(opts.ImageID),
				SourceRegion:  awsclient.String(region),
				Description:   awsclient.String(imageDesc),
				Name:          image.Name,
				DryRun:        awsclient.Boolean(opts.DryRun),
			}

			_, err := svc.CopyImage(input)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}

			wg.Done()
		}(r)
	}

	wg.Wait()
	return multiErrors
}
