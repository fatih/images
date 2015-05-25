package awsimages

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/aws/credentials"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/fatih/images/command/loader"
	"github.com/hashicorp/go-multierror"
	"github.com/shiena/ansicolor"
)

type AwsConfig struct {
	// just so we can use the Env and TOML loader more efficiently with out
	// any complex hacks
	Aws struct {
		Region        string `toml:"region" json:"region"`
		RegionExclude string `toml:"region_exclude" json:"region_exclude"`
		AccessKey     string `toml:"access_key" json:"access_key"`
		SecretKey     string `toml:"secret_key" json:"secret_key"`
	}
}

type AwsImages struct {
	services *multiRegion

	images map[string][]*ec2.Image
}

func New(args []string) *AwsImages {
	conf := new(AwsConfig)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	if conf.Aws.Region == "" {
		fmt.Fprintln(os.Stderr, "region is not set")
		os.Exit(1)
	}

	if conf.Aws.AccessKey == "" {
		fmt.Fprintln(os.Stderr, "access key is not set")
		os.Exit(1)
	}

	if conf.Aws.SecretKey == "" {
		fmt.Fprintln(os.Stderr, "secret key is not set")
		os.Exit(1)
	}

	creds := credentials.NewStaticCredentials(conf.Aws.AccessKey, conf.Aws.SecretKey, "")
	awsConfig := &aws.Config{
		Credentials: creds,
		HTTPClient:  http.DefaultClient,
		Logger:      os.Stdout,
	}

	m := newMultiRegion(awsConfig, parseRegions(conf.Aws.Region, conf.Aws.RegionExclude))
	return &AwsImages{
		services: m,
		images:   make(map[string][]*ec2.Image),
	}
}

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

		fmt.Fprintln(w, green("AWS: Region: %s (%d images):", region, len(images)))
		fmt.Fprintln(w, "    Name\tID\tState\tTags")

		for i, image := range images {
			tags := make([]string, len(image.Tags))
			for i, tag := range image.Tags {
				tags[i] = *tag.Key + ":" + *tag.Value
			}

			fmt.Fprintf(w, "[%d] %s\t%s\t%s\t%+v\n",
				i, *image.Name, *image.ImageID, *image.State, tags)
		}

		fmt.Fprintln(w, "")
	}
}

// TODO(arslan): generate dynamically, I hate writing them myself
func (a *AwsImages) Help(command string) string {
	var help string
	switch command {
	case "modify":
		help = `Usage: images modify --provider aws [options]

  Modify AMI properties.

Options:

  -image-ids   "ami-123,..."   Images to be used with below actions
  -create-tags "key=val,..."   Create or override tags
  -delete-tags "key,..."       Delete tags
  -dry-run                     Don't run command, but show the action
`
	case "delete":
		help = `Usage: images delete --provider aws [options]

  Delete (deregister) AMI images.

Options:

  -image-ids   "ami-123,..."   Images to be deleted with the given ids
  -tags        "key=val,..."   Images to be deleted with the given tags
  -dry-run                     Don't run command, but show the action
`
	case "list":
		help = `Usage: images list --provider aws [options]

  List AMI properties.

Options:
`
	default:
		return "no help found for command " + command
	}

	global := `
  -region "..."                AWS Region (env: AWS_REGION)
  -access-key "..."            AWS Access Key (env: AWS_ACCESS_KEY)
  -secret-key "..."            AWS Secret Key (env: AWS_SECRET_KEY)
`

	help += global
	return help
}

func (a *AwsImages) Delete(args []string) error {
	var (
		imageIds string
		dryRun   bool
	)

	flagSet := flag.NewFlagSet("delete", flag.ContinueOnError)
	flagSet.StringVar(&imageIds, "image-ids", "", "Images to be deleted with the given ids")
	flagSet.StringVar(&imageIds, "tags", "", "Images to be deleted with the given tags")
	flagSet.BoolVar(&dryRun, "dry-run", false, "Don't run command, but show the action")
	flagSet.Usage = func() {
		helpMsg := `Usage: images delete --provider aws [options]

  Deregister AMI's.

Options:

  -image-ids   "ami-123,..."   Images to be deleted with the given ids
  -tags        "key=val,..."   Images to be deleted with the given tags
  -dry-run                     Don't run command, but show the action
`
		fmt.Fprintf(os.Stderr, helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	if err := flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	if imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	return a.Deregister(dryRun, strings.Split(imageIds, ",")...)
}

func (a *AwsImages) Modify(args []string) error {
	var (
		createTags string
		deleteTags string
		imageIds   string
		dryRun     bool
	)

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&createTags, "create-tags", "", "Create  or override tags")
	flagSet.StringVar(&deleteTags, "delete-tags", "", "Delete tags")
	flagSet.StringVar(&imageIds, "image-ids", "", "Images to be used with actions")
	flagSet.BoolVar(&dryRun, "dry-run", false, "Don't run command, but show the action")
	flagSet.Usage = func() {
		helpMsg := `Usage: images modify --provider aws [options]

  Modify AMI properties.

Options:

  -image-ids   "ami-123,..."   Images to be used with below actions

  -create-tags "key=val,..."   Create or override tags
  -delete-tags "key,..."       Delete tags
  -dry-run                     Don't run command, but show the action
`
		fmt.Fprintf(os.Stderr, helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	if err := flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	if imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	if createTags != "" && deleteTags != "" {
		return errors.New("not allowed to be used together: [--create-tags,--delete-tags]")
	}

	if createTags != "" {
		return a.CreateTags(createTags, dryRun, strings.Split(imageIds, ",")...)
	}

	if deleteTags != "" {
		return a.DeleteTags(deleteTags, dryRun, strings.Split(imageIds, ",")...)
	}

	return nil
}

func (a *AwsImages) Deregister(dryRun bool, images ...string) error {
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

func (a *AwsImages) Copy(args []string) error {
	var (
		imageID       string
		sourceRegions string
		desc          string
		dryRun        bool
	)

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.StringVar(&imageID, "image", "", "Image to be copied with the given id")
	flagSet.StringVar(&sourceRegions, "regions", "", "Images to be copied to the given regions")
	flagSet.StringVar(&desc, "desc", "", "Description for the new AMI (optional)")
	flagSet.BoolVar(&dryRun, "dry-run", false, "Don't run command, but show the action")
	flagSet.Usage = func() {
		helpMsg := `Usage: images copy --provider aws [options]

  Copy image to regions

Options:

  -image   "ami-123"        Image to be copied with the given id
  -regions "us-east-1,..."  Image to be copied to the given regions 
  -desc    "My New Image"   Description for the new AMI's (optional)
  -dry-run                  Don't run command, but show the action
`
		fmt.Fprintf(os.Stderr, helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	if err := flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		flagSet.Usage()
		return nil
	}

	if imageID == "" {
		return errors.New("no image is passed. Use --image")
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	images, err := a.matchImages(imageID)
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
		ImageIDs: stringSlice(imageID),
	})
	if err != nil {
		return err
	}

	image := resp.Images[0]

	if desc == "" {
		desc = *image.Description
	}

	// don't all
	regions := strings.Split(sourceRegions, ",")

	for _, r := range regions {
		wg.Add(1)
		go func(region string) {
			imageDesc := fmt.Sprintf("[Copied %s from %s via images] %s", imageID, region, desc)
			log.Println("copying image ...")
			input := &ec2.CopyImageInput{
				SourceImageID: aws.String(imageID),
				SourceRegion:  aws.String(region),
				Description:   aws.String(imageDesc),
				Name:          image.Name,
				DryRun:        aws.Boolean(dryRun),
			}

			_, err := svc.CopyImage(input)
			mu.Lock()
			multiErrors = multierror.Append(multiErrors, err)
			mu.Unlock()

			wg.Done()
		}(r)
	}

	wg.Wait()
	return multiErrors
}
