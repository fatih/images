package images

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
	"text/tabwriter"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/fatih/images/command/loader"
)

type AwsConfig struct {
	// just so we can use the Env and TOML loader more efficiently with out
	// any complex hacks
	Aws struct {
		Region    string
		AccessKey string
		SecretKey string
	}
}

type AwsImages struct {
	svc *ec2.EC2

	images []*ec2.Image
}

func NewAwsImages(args []string) *AwsImages {
	conf := new(AwsConfig)
	if err := loader.Load(conf, args); err != nil {
		panic(err)
	}

	awsConfig := &aws.Config{
		Credentials: aws.DetectCreds(
			conf.Aws.AccessKey,
			conf.Aws.SecretKey,
			"",
		),
		HTTPClient: http.DefaultClient,
		Logger:     os.Stdout,
		Region:     conf.Aws.Region,
	}

	return &AwsImages{
		svc: ec2.New(awsConfig),
	}
}

func (a *AwsImages) Fetch(args []string) error {
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

  -region "..."                AWS Region (env: AWS_REGION)
  -accesskey "..."             AWS Access Key (env: AWS_ACCESS_KEY)
  -secretkey "..."             AWS Secret Key (env: AWS_SECRET_KEY)

  -image-ids   "ami-123,..."   Images to be used with below actions

  -create-tags "key=val,..."   Create or override tags
  -delete-tags "key,..."       Delete tags
  -dry-run                     Don't run command, but show the action
`
	case "list":
		return `Usage: images list --provider aws [options]

  List AMI properties.

Options:

  -region "..."                AWS Region (env: AWS_REGION)
  -accesskey "..."             AWS Access Key (env: AWS_ACCESS_KEY)
  -secretkey "..."             AWS Secret Key (env: AWS_SECRET_KEY)
`
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

// CreateTags adds or overwrites all tags for the specified images. Tags is in
// the form of "key1=val1,key2=val2,key3,key4=".
// One or more tags. The value parameter is required, but if you don't want the
// tag to have a value, specify the parameter with no value (i.e: "key3" or
// "key4=" both works)
func (a *AwsImages) CreateTags(tags string, dryRun bool, images ...string) error {
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

	params := &ec2.CreateTagsInput{
		Resources: stringSlice(images...),
		Tags:      ec2Tags,
		DryRun:    aws.Boolean(dryRun),
	}

	_, err := a.svc.CreateTags(params)
	return err
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

	_, err := a.svc.DeleteTags(params)
	return err
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
