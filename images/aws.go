package images

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
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
	)

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&createTags, "create-tags", "", "create tags")
	flagSet.StringVar(&deleteTags, "delete-tags", "", "delete tags")
	flagSet.Usage = func() {
		helpMsg := `Usage: images modify --provider aws [options]

  Modify AMI properties.

Options:

  -create-tags                  Create or override tags
  -delete-tags                  Delete tags
`
		fmt.Fprintf(os.Stderr, helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	if err := flagSet.Parse(args); err != nil {
		// flagSet.Usage()
		return nil // we don't return error, the usage will be printed instad
	}

	if len(args) == 0 {
		flagSet.Usage()
		return errors.New("no flags are passed")
	}

	fmt.Printf("createTags = %+v\n", createTags)
	fmt.Printf("deleteTags = %+v\n", deleteTags)
	return nil
}

// Add tags adds or overwrites all tags for the specified images
func (a *AwsImages) AddTags(tags map[string]string, images ...string) error {
	params := &ec2.CreateTagsInput{
		Resources: []*string{ // Required
			aws.String("String"), // Required
			// More values...
		},
		Tags: []*ec2.Tag{ // Required
			&ec2.Tag{ // Required
				Key:   aws.String("String"),
				Value: aws.String("String"),
			},
			// More values...
		},
		DryRun: aws.Boolean(true),
	}

	resp, err := a.svc.CreateTags(params)

	if awserr := aws.Error(err); awserr != nil {
		// A service error occurred.
		fmt.Println("Error:", awserr.Code, awserr.Message)
	} else if err != nil {
		// A non-service error occurred.
		panic(err)
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
