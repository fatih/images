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
		RegionExclude string `toml:"region_exclude"`
		AccessKey     string
		SecretKey     string
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

	awsConfig := &aws.Config{
		Credentials: aws.DetectCreds(
			conf.Aws.AccessKey,
			conf.Aws.SecretKey,
			"",
		),
		HTTPClient: http.DefaultClient,
		Logger:     os.Stdout,
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

// CreateTags adds or overwrites all tags for the specified images. Tags is in
// the form of "key1=val1,key2=val2,key3,key4=".
// One or more tags. The value parameter is required, but if you don't want the
// tag to have a value, specify the parameter with no value (i.e: "key3" or
// "key4=" both works)
func (a *AwsImages) CreateTags(tags string, dryRun bool, images ...string) error {
	// for one region just assume all image ids belong to the this region
	// (which `list` returns already)
	if len(a.services.regions) == 1 {
		params := &ec2.CreateTagsInput{
			Resources: stringSlice(images...),
			Tags:      populateEC2Tags(tags),
			DryRun:    aws.Boolean(dryRun),
		}

		svc, err := a.singleSvc()
		if err != nil {
			return err
		}

		_, err = svc.CreateTags(params)
		return err
	}

	// so we have multiple regions, the given images might belong to different
	// regions. Fetch all images and match each image id to the given region.
	if err := a.Fetch(nil); err != nil {
		return err
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	matchedImages := make(map[string][]string)
	for _, imageId := range images {
		region, err := a.imageRegion(imageId)
		if err != nil {
			multiErrors = multierror.Append(multiErrors, err)
			continue
		}

		ids := matchedImages[region]
		ids = append(ids, imageId)
		matchedImages[region] = ids
	}

	// return early if we have any error while checking the ids
	if multiErrors != nil {
		return multiErrors
	}

	ec2Tags := populateEC2Tags(tags)

	for r, i := range matchedImages {
		wg.Add(1)
		go func(region string, images []string) {
			defer wg.Done()

			params := &ec2.CreateTagsInput{
				Resources: stringSlice(images...),
				Tags:      ec2Tags,
				DryRun:    aws.Boolean(dryRun),
			}

			svc, err := a.svcFromRegion(region)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
				return
			}

			_, err = svc.CreateTags(params)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
		}(r, i)
	}
	wg.Wait()

	return multiErrors
}

func (a *AwsImages) imageRegion(imageId string) (string, error) {
	for region, images := range a.images {
		for _, image := range images {
			if *image.ImageID == imageId {
				return region, nil
			}
		}
	}

	return "", fmt.Errorf("no region found for image id '%s'", imageId)
}

func populateEC2Tags(tags string) []*ec2.Tag {
	ec2Tags := make([]*ec2.Tag, 0)
	for _, keyVal := range strings.Split(tags, ",") {
		keys := strings.Split(keyVal, "=")
		ec2Tag := &ec2.Tag{
			Key: aws.String(keys[0]), // index 0 is always available
		}

		// It's in the form "key4". The AWS API will create the key only if the
		// value is being passed as an empty string.
		if len(keys) == 1 {
			ec2Tag.Value = aws.String("")
		}

		if len(keys) == 2 {
			ec2Tag.Value = aws.String(keys[1])
		}

		ec2Tags = append(ec2Tags, ec2Tag)
	}

	return ec2Tags
}

// DeleteTags deletes the given tags for the given images. Tags is in the form
// of "key1=val1,key2=val2,key3,key4="
// One or more tags to delete. If you omit the value parameter(i.e "key3"), we
// delete the tag regardless of its value. If you specify this parameter with
// an empty string (i.e: "key4=" as the value, we delete the key only if its
// value is an empty string.
func (a *AwsImages) DeleteTags(tags string, dryRun bool, images ...string) error {
	ec2Tags := make([]*ec2.Tag, 0)

	for _, keyVal := range strings.Split(tags, ",") {
		keys := strings.Split(keyVal, "=")
		ec2Tag := &ec2.Tag{
			Key: aws.String(keys[0]), // index 0 is always available
		}

		// means value is not omitted. We don't care if value is empty or not,
		// the AWS API takes care of it.
		if len(keys) == 2 {
			ec2Tag.Value = aws.String(keys[1])
		}

		ec2Tags = append(ec2Tags, ec2Tag)
	}

	params := &ec2.DeleteTagsInput{
		Resources: stringSlice(images...),
		Tags:      ec2Tags,
		DryRun:    aws.Boolean(dryRun),
	}

	svc, err := a.singleSvc()
	if err != nil {
		return err
	}

	_, err = svc.DeleteTags(params)
	return err
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

func stringSlice(vals ...string) []*string {
	a := make([]*string, len(vals))

	for i, v := range vals {
		a[i] = aws.String(v)
	}

	return a
}
