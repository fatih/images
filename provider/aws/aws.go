package aws

import (
	"errors"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	awsclient "github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/aws/credentials"
	"github.com/awslabs/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
)

type AwsConfig struct {
	Regions        []string `toml:"regions" json:"regions"`
	RegionsExclude []string `toml:"regions_exclude" json:"regions_exclude"`
	AccessKey      string   `toml:"access_key" json:"access_key"`
	SecretKey      string   `toml:"secret_key" json:"secret_key"`
}

// AwsImages is responsible of managing AWS images (AMI's)
type AwsImages struct {
	services *multiRegion
	images   Images
}

func New(conf *AwsConfig) (*AwsImages, error) {
	checkCfg := "Please check your configuration"

	if len(conf.Regions) == 0 {
		return nil, errors.New("AWS Regions are not set. " + checkCfg)
	}

	if conf.AccessKey == "" {
		return nil, errors.New("AWS Access Key is not set. " + checkCfg)
	}

	if conf.SecretKey == "" {
		return nil, errors.New("AWS Secret Key is not set. " + checkCfg)
	}

	// increase the timeout
	timeout := time.Second * 30
	client := &http.Client{
		Transport: &http.Transport{TLSHandshakeTimeout: timeout},
		Timeout:   timeout,
	}

	creds := credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, "")
	awsCfg := &awsclient.Config{
		Credentials: creds,
		HTTPClient:  client,
		Logger:      os.Stdout,
	}

	m := newMultiRegion(awsCfg, filterRegions(conf.Regions, conf.RegionsExclude))
	return &AwsImages{
		services: m,
		images:   make(map[string][]*ec2.Image),
	}, nil
}

func (a *AwsImages) Images(input *ec2.DescribeImagesInput) (Images, error) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex

		multiErrors error
	)

	images := make(map[string][]*ec2.Image)

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

				images[region] = resp.Images
			}

			mu.Unlock()
			wg.Done()
		}(r, s)
	}

	wg.Wait()

	return images, multiErrors
}

func (a *AwsImages) ownerImages() (Images, error) {
	input := &ec2.DescribeImagesInput{
		Owners: stringSlice("self"),
	}

	return a.Images(input)
}
