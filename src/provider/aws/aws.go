package aws

import (
	"errors"
	"net/http"
	"time"

	awsclient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
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
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: timeout,
		},
		Timeout: timeout,
	}

	creds := credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, "")
	awsCfg := &awsclient.Config{
		Credentials: creds,
		HTTPClient:  client,
		Logger:      awsclient.NewDefaultLogger(),
	}

	m := newMultiRegion(awsCfg, filterRegions(conf.Regions, conf.RegionsExclude))
	return &AwsImages{
		services: m,
		images:   make(map[string][]*ec2.Image),
	}, nil
}
