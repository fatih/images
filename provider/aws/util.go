package aws

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	awsclient "github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

type multiFunc func(svc *ec2.EC2, images []string) error

// multiCall calls the given function concurrently for each region. It fetches
// and matches the region to each image automatically.
func (a *AwsImages) multiCall(fn multiFunc, images ...string) error {
	// for one region just assume all image ids belong to the this region
	// (which `list` returns already)
	if len(a.services.regions) == 1 {
		svc, err := a.singleSvc()
		if err != nil {
			return err
		}

		return fn(svc, images)
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

			if err := fn(svc, images); err != nil {
				mu.Lock()
				multiErrors = multierror.Append(multiErrors, err)
				mu.Unlock()
			}
		}(r, i)
	}

	wg.Wait()
	return multiErrors

}

// matchImages matches the given images to their respective regions and returns
// map of region to images.
func (a *AwsImages) matchImages(images ...string) (map[string][]string, error) {
	ownerImages, err := a.ownerImages()
	if err != nil {
		return nil, err
	}

	matchedImages := make(map[string][]string)
	for _, imageID := range images {
		region, err := ownerImages.RegionFromId(imageID)
		if err != nil {
			return nil, err
		}

		ids := matchedImages[region]
		ids = append(ids, imageID)
		matchedImages[region] = ids
	}

	return matchedImages, nil
}

// singleSvc returns a single *ec2.EC2 service from the list of regions.
func (a *AwsImages) singleSvc() (*ec2.EC2, error) {
	if len(a.services.regions) > 1 {
		return nil, errors.New("multiple regions are available for singleSvc")
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

	return nil, fmt.Errorf("no svc found for region '%s'", region)
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
		a[i] = awsclient.String(v)
	}

	return a
}
