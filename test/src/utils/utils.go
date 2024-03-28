package utils

import (
	"context"
	"encoding/base64"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	types2 "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func NewClientSet(cluster *types.Cluster) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: *cluster.Name,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        *cluster.Endpoint,
			BearerToken: tok.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: ca,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func GetAwsClient() (aws.Config, error) {
	awsProfile := GetEnv("AWS_PROFILE", GetEnv("AWS_DEFAULT_PROFILE", "infex"))
	region := GetEnv("AWS_REGION", "eu-central-1")

	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(awsProfile),
	)
}

// CheckVpcAttribute Function to check a specific VPC attribute
func CheckVpcAttribute(ec2Svc *ec2.Client, vpcID string, attributeName types2.VpcAttributeName) (*ec2.DescribeVpcAttributeOutput, error) {
	input := &ec2.DescribeVpcAttributeInput{
		VpcId:     &vpcID,
		Attribute: attributeName,
	}

	return ec2Svc.DescribeVpcAttribute(context.TODO(), input)
}
