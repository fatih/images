package awsimages

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type DeleteOptions struct {
	// Images to be deleted
	ImageIds []string

	// DryRun doesn't run the command, but shows the action
	DryRun bool

	helpMsg string
	flagSet *flag.FlagSet
}

func newDeleteOptions() *DeleteOptions {
	d := &DeleteOptions{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.Var(flags.StringListVar(&d.ImageIds), "ids", "Images to be delete with the given ids")
	flagSet.BoolVar(&d.DryRun, "dry-run", false, "Don't run command, but show the action")
	d.helpMsg = `Usage: images delete --provider aws [options]

  Deregister AMI's.

Options:

  -ids         "ami-123,..."   Images to be deleted with the given ids
  -dry-run                     Don't run command, but show the action
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, d.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	d.flagSet = flagSet
	return d
}

// Delete deletes the given images.
func (a *AwsImages) DeleteImages(opts *DeleteOptions) error {
	deleteImages := func(svc *ec2.EC2, images []string) error {
		var multiErrors error

		for _, image := range images {
			input := &ec2.DeregisterImageInput{
				ImageID: aws.String(image),
				DryRun:  aws.Boolean(opts.DryRun),
			}

			_, err := svc.DeregisterImage(input)
			if err != nil {
				multiErrors = multierror.Append(multiErrors, err)
			}
		}

		return multiErrors
	}

	return a.multiCall(deleteImages, opts.ImageIds...)
}
