module github.com/camunda/camunda-tf-eks-module

go 1.22.5

require (
	github.com/aws/aws-sdk-go-v2/service/s3 v1.58.2
	github.com/aws/aws-sdk-go-v2 v1.30.4
	github.com/aws/aws-sdk-go-v2/config v1.27.28
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.175.1
	github.com/aws/aws-sdk-go-v2/service/eks v1.48.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.35.0
	github.com/aws/aws-sdk-go-v2/service/kms v1.35.4
	github.com/aws/aws-sdk-go-v2/service/rds v1.82.1
	github.com/aws/smithy-go v1.20.4
	github.com/gruntwork-io/terratest v0.47.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	k8s.io/api v0.31.0
	k8s.io/apimachinery v0.31.0
	k8s.io/client-go v0.31.0
	sigs.k8s.io/aws-iam-authenticator v0.6.23
)
