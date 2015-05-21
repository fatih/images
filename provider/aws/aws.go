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
	"time"

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
		Region        string
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
  -accesskey "..."             AWS Access Key (env: AWS_ACCESS_KEY)
  -secretkey "..."             AWS Secret Key (env: AWS_SECRET_KEY)
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

// singleSvc returns a single *ec2.EC2 service from the list of regions.
func (a *AwsImages) singleSvc() (*ec2.EC2, error) {
	if len(a.services.regions) > 1 {
		return nil, errors.New("deleting images for multiple regions is not supported")
	}

	var svc *ec2.EC2
	for _, s := range a.services.regions {
		svc = s
	}

	return svc, nil
}

// svcFromRegion returns a *ec2.EC2 service with the given region
func (a *AwsImages) svcFromRegion(region string) (*ec2.EC2, error) {
	for r, s := range a.services.regions {
		if r == region {
			return s, nil
		}
	}

	return nil, fmt.Errorf("no svc found for region '%s'")
}

func (a *AwsImages) Deregister(dryRun bool, images ...string) error {
	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	svc, err := a.singleSvc()
	if err != nil {
		return err
	}

	for _, imageId := range images {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			input := &ec2.DeregisterImageInput{
				ImageID: aws.String(imageId),
				DryRun:  aws.Boolean(dryRun),
			}

			_, err := svc.DeregisterImage(input)
			mu.Lock()
			multiErrors = multierror.Append(multiErrors, err)
			mu.Unlock()
		}(imageId)
	}

	wg.Wait()
	return multiErrors
}

func (a *AwsImages) Copy(args []string) error {
	var (
		imageIds string
		regions  string
		desc     string
		dryRun   bool
	)

	flagSet := flag.NewFlagSet("copy", flag.ContinueOnError)
	flagSet.StringVar(&imageIds, "image-ids", "", "Images to be copied with the given ids")
	flagSet.StringVar(&regions, "regions", "", "Images to be copied to the given regions")
	flagSet.StringVar(&desc, "description", "", "Description for the new AMI")
	flagSet.BoolVar(&dryRun, "dry-run", false, "Don't run command, but show the action")
	flagSet.Usage = func() {
		helpMsg := `Usage: images copy --provider aws [options]

  Deregister AMI's.

Options:

  -image-ids   "ami-123,..."   Images to be copied with the given ids
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

	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	svc, err := a.singleSvc()
	if err != nil {
		return err
	}

	images := strings.Split(imageIds, ",")

	for _, imageId := range images {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			input := &ec2.CopyImageInput{
				SourceImageID: aws.String(imageId),
				DryRun:        aws.Boolean(dryRun),
			}

			_, err := svc.CopyImage(input)
			mu.Lock()
			multiErrors = multierror.Append(multiErrors, err)
			mu.Unlock()
		}(imageId)
	}

	wg.Wait()
	return multiErrors
}

// imageRegion returns the given imageId's region
func (a *AwsImages) imageRegion(imageId string) (string, error) {
	if len(a.images) == 0 {
		return "", errors.New("images are not fetched")
	}

	for region, images := range a.images {
		for _, image := range images {
			if *image.ImageID == imageId {
				return region, nil
			}
		}
	}

	return "", fmt.Errorf("no region found for image id '%s'", imageId)
}

// matchImages matches the given images to their respective regions and returns
// map of region to images.
func (a *AwsImages) matchImages(images ...string) (map[string][]string, error) {
	if err := a.Fetch(nil); err != nil {
		return nil, err
	}

	matchedImages := make(map[string][]string)
	for _, imageId := range images {
		region, err := a.imageRegion(imageId)
		if err != nil {
			return nil, err
		}

		ids := matchedImages[region]
		ids = append(ids, imageId)
		matchedImages[region] = ids
	}

	return matchedImages, nil
}

// byTime implements sort.Interface for []*ec2.Image based on the CreationDate field.
type byTime []*ec2.Image

func (a byTime) Len() int      { return len(a) }
func (a byTime) Swap(i, j int) { *a[i], *a[j] = *a[j], *a[i] }
func (a byTime) Less(i, j int) bool {
	it, err := time.Parse(time.RFC3339, *a[i].CreationDate)
	if err != nil {
		log.Println("aws: sorting err: ", err)
	}

	jt, err := time.Parse(time.RFC3339, *a[j].CreationDate)
	if err != nil {
		log.Println("aws: sorting err: ", err)
	}

	return it.Before(jt)
}

// stringSlice is an helper method to convert a slice of strings into a slice
// of pointer of strings. Needed for various aws/ec2 commands.
func stringSlice(vals ...string) []*string {
	a := make([]*string, len(vals))

	for i, v := range vals {
		a[i] = aws.String(v)
	}

	return a
}
