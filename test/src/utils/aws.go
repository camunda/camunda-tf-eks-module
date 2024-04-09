package utils

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"time"
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

func WaitForUpdateEKS(ctx context.Context, client *eks.Client, clusterName, updateID string) error {
	describeUpdateInput := &eks.DescribeUpdateInput{
		Name:     &clusterName,
		UpdateId: &updateID,
	}

L:
	for {
		updateOutput, err := client.DescribeUpdate(ctx, describeUpdateInput)
		if err != nil {
			return err
		}

		status := updateOutput.Update.Status
		fmt.Printf("Update status: %s\n", status)

		switch status {
		case types.UpdateStatusFailed:
			return fmt.Errorf("update failed")
		case types.UpdateStatusCancelled:
			return fmt.Errorf("update cancelled")
		case types.UpdateStatusSuccessful:
			break L
		case types.UpdateStatusInProgress:
			time.Sleep(5 * time.Second)
		default:
			return fmt.Errorf("update status unknown: %s", status)
		}
	}

	return nil
}

func UpgradeEKS(ctx context.Context, client *eks.Client, clusterName, version string) error {
	input := &eks.UpdateClusterVersionInput{
		Name:    &clusterName,
		Version: &version,
	}

	output, err := client.UpdateClusterVersion(ctx, input)
	if err != nil {
		return err
	}

	fmt.Printf("Update initiated, update ID: %s\n", *output.Update.Id)

	err = WaitForUpdateEKS(ctx, client, clusterName, *output.Update.Id)
	if err != nil {
		return err
	}

	fmt.Println("Update completed successfully")
	return nil
}
