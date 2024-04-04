package utils

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func GetAwsProfile() string {
	return GetEnv("AWS_PROFILE", GetEnv("AWS_DEFAULT_PROFILE", "infex"))
}

func GetAwsRegion() string {
	return GetEnv("AWS_REGION", "eu-central-1")
}

// GetAwsClient returns an aws.Config client from the env variables `AWS_PROFILE` and `AWS_REGION`
func GetAwsClient() (aws.Config, error) {
	awsProfile := GetAwsProfile()
	region := GetAwsRegion()

	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(awsProfile),
	)
}
