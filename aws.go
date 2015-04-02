package images

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/awslabs/aws-sdk-go/aws"
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

	a.images = make([]*ec2.Image, len(resp.Images))
	for i, image := range resp.Images {
		a.images[i] = image
	}

	return err
}

func (a *AwsImages) Print() {
	if len(a.images) == 0 {
		fmt.Println("no images found")
		return
	}

	color.Green("AWS: %d images found\n\n", len(a.images))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 10, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintln(w, "    Name\tID\tTags")
	for i, image := range a.images {
		tags := make([]string, len(image.Tags))
		for i, tag := range image.Tags {
			tags[i] = *tag.Key + ":" + *tag.Value
		}

		fmt.Fprintf(w, "[%d] %s\t%s\t%+v\n", i, *image.Name, *image.ImageID, tags)
	}
}

func stringSlice(vals ...string) []*string {
	a := make([]*string, len(vals))

	for i, v := range vals {
		a[i] = aws.String(v)
	}

	return a
}
