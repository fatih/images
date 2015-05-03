package awsimages

import (
	"strings"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/ec2"
)

var allRegions = []string{
	"ap-northeast-1",
	"ap-southeast-1",
	"ap-southeast-2",
	"cn-north-1",
	"eu-central-1",
	"eu-west-1",
	"sa-east-1",
	"us-east-1",
	"us-gov-west-1",
	"us-west-1",
	"us-west-2",
}

type multiRegion struct {
	regions map[string]*ec2.EC2
}

func newMultiRegion(conf *aws.Config, regions []string) *multiRegion {
	m := &multiRegion{
		regions: make(map[string]*ec2.EC2, 0),
	}

	for _, region := range regions {
		m.regions[region] = ec2.New(conf.Merge(&aws.Config{Region: region}))
	}

	return m
}

func parseRegions(region, exclude string) []string {
	regions := strings.Split(region, ",")
	if region == "all" {
		regions = allRegions
	}

	excludedRegions := strings.Split(exclude, ",")

	inExcluded := func(r string) bool {
		for _, region := range excludedRegions {
			if r == region {
				return true
			}
		}
		return false
	}

	finalRegions := make([]string, 0)
	for _, r := range regions {
		if !inExcluded(r) {
			finalRegions = append(finalRegions, r)
		}
	}

	return finalRegions
}
