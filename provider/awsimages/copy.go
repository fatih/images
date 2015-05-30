package awsimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

type copyFlags struct {
	imageID       string
	sourceRegions string
	desc          string
	dryRun        bool
	helpMsg       string

	flagSet *flag.FlagSet
}

func newCopyFlags() *copyFlags {
	c := &copyFlags{}

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.StringVar(&c.imageID, "image", "", "Image to be copied with the given id")
	flagSet.StringVar(&c.sourceRegions, "to", "", "Images to be copied to the given regions")
	flagSet.StringVar(&c.desc, "desc", "", "Description for the new AMI (optional)")
	flagSet.BoolVar(&c.dryRun, "dry-run", false, "Don't run command, but show the action")

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

func (a *AwsImages) Copy(args []string) error {
	c := newCopyFlags()

	if err := c.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		c.flagSet.Usage()
		return nil
	}

	if c.imageID == "" {
		return errors.New("no image is passed. Use --image")
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	images, err := a.matchImages(c.imageID)
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
		ImageIDs: stringSlice(c.imageID),
	})
	if err != nil {
		return err
	}

	image := resp.Images[0]

	if c.desc == "" {
		c.desc = *image.Description
	}

	regions := strings.Split(c.sourceRegions, ",")

	for _, r := range regions {
		wg.Add(1)
		go func(region string) {
			imageDesc := fmt.Sprintf("[Copied %s from %s via images] %s", c.imageID, region, c.desc)
			log.Println("copying image ...")
			input := &ec2.CopyImageInput{
				SourceImageID: aws.String(c.imageID),
				SourceRegion:  aws.String(region),
				Description:   aws.String(imageDesc),
				Name:          image.Name,
				DryRun:        aws.Boolean(c.dryRun),
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
