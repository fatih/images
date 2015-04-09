package images

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/aws/awsutil"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
)

type AwsImages struct {
	svc *ec2.EC2

	images []*ec2.Image
}

func NewAwsImages(region string) *AwsImages {
	return &AwsImages{
		svc: ec2.New(&aws.Config{Region: region}),
	}
}

func (a *AwsImages) Fetch() error {
	input := &ec2.DescribeImagesInput{
		Owners: stringSlice("self"),
	}

	resp, err := a.svc.DescribeImages(input)
	if err != nil {
		return err
	}

	a.images = resp.Images

	// sort from oldest to newest
	if len(a.images) > 1 {
		sort.Sort(a)
	}

	return nil
}

func (a *AwsImages) Print() {
	if len(a.images) == 0 {
		fmt.Println("no images found")
		return
	}

	color.Green("AWS: Region: %s (%d images)\n\n", a.svc.Config.Region, len(a.images))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 10, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintln(w, "    Name\tID\tState\tTags")

	for i, image := range a.images {
		tags := make([]string, len(image.Tags))
		for i, tag := range image.Tags {
			tags[i] = *tag.Key + ":" + *tag.Value
		}

		fmt.Fprintf(w, "[%d] %s\t%s\t%s\t%+v\n",
			i, *image.Name, *image.ImageID, *image.State, tags)
	}
}

func (a *AwsImages) Help(command string) string {
	switch command {
	case "modify":
		return `Usage: images modify --provider aws [options] 

  Modify AMI properties. 

Options:

  -create-tags                  Create or override tags
  -delete-tags                  Delete tags
`
	case "list":
	}

	return "no help found for command " + command
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
		return errors.New("no flags are passed")
	}

	if imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	if createTags != "" && deleteTags != "" {
		return errors.New("not allowed to be used together: [--create-tags,--delete-tags]")
	}
	fmt.Printf("imageIds = %+v\n", imageIds)

	if createTags != "" {
		fmt.Printf("createTags = %+v\n", createTags)

		keyVals := make(map[string]string, 0)

		for _, keyVal := range strings.Split(createTags, ",") {
			keys := strings.Split(keyVal, "=")
			if len(keys) != 2 {
				return fmt.Errorf("malformed value passed to --create-tags: %v", keys)
			}
			keyVals[keys[0]] = keys[1]
		}

		return a.AddTags(keyVals, dryRun, strings.Split(imageIds, ",")...)
	}

	if deleteTags != "" {
		fmt.Printf("deleteTags = %+v\n", deleteTags)
	}

	return nil
}

// Add tags adds or overwrites all tags for the specified images
func (a *AwsImages) AddTags(tags map[string]string, dryRun bool, images ...string) error {
	ec2Tags := make([]*ec2.Tag, 0)
	for key, val := range tags {
		ec2Tags = append(ec2Tags, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}

	params := &ec2.CreateTagsInput{
		Resources: stringSlice(images...),
		Tags:      ec2Tags,
		DryRun:    aws.Boolean(dryRun),
	}

	resp, err := a.svc.CreateTags(params)
	if awserr := aws.Error(err); awserr != nil {
		// A service error occurred.
		// fmt.Println("Error:", awserr.Code, awserr.Message)
		return err
	} else if err != nil {
		// A non-service error occurred.
		return err
	}

	// Pretty-print the response data.
	fmt.Println(awsutil.StringValue(resp))
	return nil
}

//
// Sort interface
//

func (a *AwsImages) Len() int {
	return len(a.images)
}

func (a *AwsImages) Less(i, j int) bool {
	it, err := time.Parse(time.RFC3339, *a.images[i].CreationDate)
	if err != nil {
		log.Println("aws: sorting err: ", err)
	}

	jt, err := time.Parse(time.RFC3339, *a.images[j].CreationDate)
	if err != nil {
		log.Println("aws: sorting err: ", err)
	}

	return it.Before(jt)
}

func (a *AwsImages) Swap(i, j int) {
	a.images[i], a.images[j] = a.images[j], a.images[i]
}

//
// Utils
//

func stringSlice(vals ...string) []*string {
	a := make([]*string, len(vals))

	for i, v := range vals {
		a[i] = aws.String(v)
	}

	return a
}
