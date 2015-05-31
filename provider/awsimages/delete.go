package awsimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

type deleteFlags struct {
	imageIds string
	dryRun   bool
	helpMsg  string

	flagSet *flag.FlagSet
}

func newDeleteFlags() *deleteFlags {
	d := &deleteFlags{}

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.StringVar(&d.imageIds, "ids", "", "Images to be deleted with the given ids")
	flagSet.StringVar(&d.imageIds, "tags", "", "Images to be deleted with the given tags")
	flagSet.BoolVar(&d.dryRun, "dry-run", false, "Don't run command, but show the action")
	d.helpMsg = `Usage: images delete --provider aws [options]

  Deregister AMI's.

Options:

  -ids         "ami-123,..."   Images to be deleted with the given ids
  -tags        "key=val,..."   Images to be deleted with the given tags
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
func (a *AwsImages) Delete(args []string) error {
	d := newDeleteFlags()

	if err := d.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		d.flagSet.Usage()
		return nil
	}

	if d.imageIds == "" {
		return errors.New("no images are passed with [--ids]")
	}

	return a.deregister(d.dryRun, strings.Split(d.imageIds, ",")...)
}

func (a *AwsImages) deregister(dryRun bool, images ...string) error {
	deleteImages := func(svc *ec2.EC2, images []string) error {
		var multiErrors error

		for _, image := range images {
			input := &ec2.DeregisterImageInput{
				ImageID: aws.String(image),
				DryRun:  aws.Boolean(dryRun),
			}

			_, err := svc.DeregisterImage(input)
			if err != nil {
				multiErrors = multierror.Append(multiErrors, err)
			}
		}

		return multiErrors
	}

	return a.multiCall(deleteImages, images...)
}
