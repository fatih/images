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
)

type modifyFlags struct {
	createTags string
	deleteTags string
	imageIds   string
	dryRun     bool
	helpMsg    string

	flagSet *flag.FlagSet
}

func newModifyFlags() *modifyFlags {
	m := &modifyFlags{}

	flagSet := flag.NewFlagSet("modify", flag.ContinueOnError)
	flagSet.StringVar(&m.createTags, "create-tags", "", "Create  or override tags")
	flagSet.StringVar(&m.deleteTags, "delete-tags", "", "Delete tags")
	flagSet.StringVar(&m.imageIds, "image-ids", "", "Images to be used with actions")
	flagSet.BoolVar(&m.dryRun, "dry-run", false, "Don't run command, but show the action")
	m.helpMsg = `Usage: images modify --provider aws [options]

  Modify AMI properties.

Options:

  -image-ids   "ami-123,..."   Images to be used with below actions

  -create-tags "key=val,..."   Create or override tags
  -delete-tags "key,..."       Delete tags
  -dry-run                     Don't run command, but show the action
`
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, m.helpMsg)
	}

	flagSet.SetOutput(ioutil.Discard) // don't print anything without my permission
	m.flagSet = flagSet
	return m
}

func (a *AwsImages) Modify(args []string) error {
	m := newModifyFlags()
	if err := m.flagSet.Parse(args); err != nil {
		return nil // we don't return error, the usage will be printed instead
	}

	if len(args) == 0 {
		m.flagSet.Usage()
		return nil
	}

	if m.imageIds == "" {
		return errors.New("no images are passed with [--image-ids]")
	}

	if m.createTags != "" && m.deleteTags != "" {
		return errors.New("not allowed to be used together: [--create-tags,--delete-tags]")
	}

	if m.createTags != "" {
		return a.CreateTags(m.createTags, m.dryRun, strings.Split(m.imageIds, ",")...)
	}

	if m.deleteTags != "" {
		return a.DeleteTags(m.deleteTags, m.dryRun, strings.Split(m.imageIds, ",")...)
	}

	return nil
}

// CreateTags adds or overwrites all tags for the specified images. Tags is in
// the form of "key1=val1,key2=val2,key3,key4=".
// One or more tags. The value parameter is required, but if you don't want the
// tag to have a value, specify the parameter with no value (i.e: "key3" or
// "key4=" both works)
func (a *AwsImages) CreateTags(tags string, dryRun bool, images ...string) error {
	createTags := func(svc *ec2.EC2, images []string) error {
		_, err := svc.CreateTags(&ec2.CreateTagsInput{
			Resources: stringSlice(images...),
			Tags:      populateEC2Tags(tags, true),
			DryRun:    aws.Boolean(dryRun),
		})
		return err
	}

	return a.multiCall(createTags, images...)
}

// DeleteTags deletes the given tags for the given images. Tags is in the form
// of "key1=val1,key2=val2,key3,key4="
// One or more tags to delete. If you omit the value parameter(i.e "key3"), we
// delete the tag regardless of its value. If you specify this parameter with
// an empty string (i.e: "key4=" as the value, we delete the key only if its
// value is an empty string.
func (a *AwsImages) DeleteTags(tags string, dryRun bool, images ...string) error {
	deleteTags := func(svc *ec2.EC2, images []string) error {
		_, err := svc.DeleteTags(&ec2.DeleteTagsInput{
			Resources: stringSlice(images...),
			Tags:      populateEC2Tags(tags, false),
			DryRun:    aws.Boolean(dryRun),
		})
		return err
	}

	return a.multiCall(deleteTags, images...)
}

// populateEC2Tags returns a list of *ec2.Tag. tags is in the form of
// "key1=val1,key2=val2,key3,key4="
func populateEC2Tags(tags string, create bool) []*ec2.Tag {
	ec2Tags := make([]*ec2.Tag, 0)
	for _, keyVal := range strings.Split(tags, ",") {
		keys := strings.Split(keyVal, "=")
		ec2Tag := &ec2.Tag{
			Key: aws.String(keys[0]), // index 0 is always available
		}

		// It's in the form "key4". The AWS API will create the key only if the
		// value is being passed as an empty string.
		if create && len(keys) == 1 {
			ec2Tag.Value = aws.String("")
		}

		if len(keys) == 2 {
			ec2Tag.Value = aws.String(keys[1])
		}

		ec2Tags = append(ec2Tags, ec2Tag)
	}

	return ec2Tags
}
