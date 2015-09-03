package aws

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"

	"provider/utils"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/flags"
	"github.com/hashicorp/go-multierror"
)

type listFlags struct {
	output   utils.OutputMode
	imageIds []string
	owners   []string

	helpMsg string
	flagSet *flag.FlagSet
}

func newListFlags() *listFlags {
	l := &listFlags{}

	flagSet := flag.NewFlagSet("list", flag.ContinueOnError)
	flagSet.Var(utils.NewOutputValue(utils.Simplified, &l.output), "output", "Output mode")
	flagSet.Var(flags.NewStringSlice(nil, &l.imageIds), "ids", "Images to be listed. Default case is all images")
	flagSet.Var(flags.NewStringSlice(nil, &l.owners), "owners", "Filters the images by the owner. By ddefault self is being used")
	l.helpMsg = `Usage: images list --providers aws [options]

   List AMI properties.

Options:

  -ids     "ami-123,..."       Images to be listed. By default all images are shown.
  -owners  "self,..."          Filters the images by the owner. By default self is being used.
  -output  "json"              Output mode of images. (default: "simplified")
                               Available options: "json","table" or "simplified" 
`

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, l.helpMsg)
	}
	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	l.flagSet = flagSet
	return l
}

func (a *AwsImages) Images(input *ec2.DescribeImagesInput) (Images, error) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	images := make(map[string][]*ec2.Image)

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

				images[region] = resp.Images
			}

			mu.Unlock()
			wg.Done()
		}(r, s)
	}

	wg.Wait()

	return images, multiErrors
}

func (a *AwsImages) ownerImages() (Images, error) {
	input := &ec2.DescribeImagesInput{
		Owners: stringSlice("self"),
	}

	return a.Images(input)
}
