package awsimages

import (
	"strings"
	"sync"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

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
	// for one region just assume all image ids belong to the this region
	// (which `list` returns already)
	if len(a.services.regions) == 1 {
		svc, err := a.singleSvc()
		if err != nil {
			return err
		}

		return createTags(svc, images)
	}

	// so we have multiple regions, the given images might belong to different
	// regions. Fetch all images and match each image id to the given region.
	matchedImages, err := a.matchImages(images...)
	if err != nil {
		return err
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for r, i := range matchedImages {
		wg.Add(1)
		go func(region string, images []string) {
			defer wg.Done()

			svc, err := a.svcFromRegion(region)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
				return
			}

			if err := createTags(svc, images); err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
		}(r, i)
	}

	wg.Wait()

	return multiErrors
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

	// for one region just assume all image ids belong to the this region
	// (which `list` returns already)
	if len(a.services.regions) == 1 {
		svc, err := a.singleSvc()
		if err != nil {
			return err
		}

		return deleteTags(svc, images)
	}

	// so we have multiple regions, the given images might belong to different
	// regions. Fetch all images and match each image id to the given region.
	matchedImages, err := a.matchImages(images...)
	if err != nil {
		return err
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex // protects multiErrors
		multiErrors error
	)

	for r, i := range matchedImages {
		wg.Add(1)
		go func(region string, images []string) {
			defer wg.Done()

			svc, err := a.svcFromRegion(region)
			if err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
				return
			}

			if err := deleteTags(svc, images); err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
		}(r, i)
	}

	wg.Wait()

	return multiErrors
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
