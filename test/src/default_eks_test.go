package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/camunda/camunda-tf-eks-module/utils"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type DefaultEKSTestSuite struct {
	suite.Suite
	logger         *zap.Logger
	sugaredLogger  *zap.SugaredLogger
	clusterName    string
	expectedNodes  int
	kubeConfigPath string
	region         string
	tfDataDir      string
	varTf          map[string]interface{}
}

func (suite *DefaultEKSTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.sugaredLogger = suite.logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	suite.clusterName = fmt.Sprintf("cluster-test-%s", clusterSuffix)
	suite.region = utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	suite.expectedNodes = 4
	var errAbsPath error
	suite.tfDataDir, errAbsPath = filepath.Abs(fmt.Sprintf("../../test/states/tf-data-%s", suite.clusterName))
	suite.Require().NoError(errAbsPath)
	suite.kubeConfigPath = fmt.Sprintf("%s/kubeconfig-default-eks", suite.tfDataDir)
}

func (suite *DefaultEKSTestSuite) TearUpTest() {
	// create tf state
	absPath, err := filepath.Abs(suite.tfDataDir)
	suite.Require().NoError(err)
	err = os.MkdirAll(absPath, os.ModePerm)
	suite.Require().NoError(err)
}

func (suite *DefaultEKSTestSuite) TearDownTest() {
	suite.T().Log("Cleaning up resources...")

	err := os.Remove(suite.kubeConfigPath)
	if err != nil && !os.IsNotExist(err) {
		suite.T().Errorf("Failed to remove kubeConfigPath: %v", err)
	}
}

// TestDefaultEKS spawns an EKS cluster with the default parameters and checks the parameters
func (suite *DefaultEKSTestSuite) TestDefaultEKS() {

	suite.varTf = map[string]interface{}{
		"name":                  suite.clusterName,
		"region":                suite.region,
		"np_desired_node_count": suite.expectedNodes,
	}

	fullDir := fmt.Sprintf("%s/eks-cluster/", suite.tfDataDir)
	errTfDir := os.MkdirAll(fullDir, os.ModePerm)
	suite.Require().NoError(errTfDir)

	tfDir := test_structure.CopyTerraformFolderToDest(suite.T(), "../../modules/", "eks-cluster/", fullDir)

	terraformOptions := &terraform.Options{
		TerraformDir: tfDir,
		Upgrade:      false,
		VarFiles:     []string{"../fixtures/fixtures.default.eks.tfvars"},
		Vars:         suite.varTf,
	}

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")

	if cleanClusterAtTheEnd == "true" {
		defer terraform.Destroy(suite.T(), terraformOptions)
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(suite.T(), terraformOptions)
		})
	}

	terraform.InitAndApplyAndIdempotent(suite.T(), terraformOptions)
	suite.baseChecksEKS(terraformOptions)
}

// baseChecksEKS checks the defaults of an EKS cluster
func (suite *DefaultEKSTestSuite) baseChecksEKS(terraformOptions *terraform.Options) {
	clusterName := terraformOptions.Vars["name"].(string)
	suite.sugaredLogger.Infow("Testing status of the EKS cluster", "clusterName", clusterName)

	// Do some basic not empty tests on outputs
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "cluster_endpoint"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "cluster_security_group_id"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "cluster_security_group_arn"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "cluster_primary_security_group_id"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "cluster_iam_role_arn"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "ebs_cs_arn"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "external_dns_arn"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "vpc_id"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "private_vpc_cidr_blocks"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "private_subnet_ids"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "default_security_group_id"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "vpc_main_route_table_id"))
	suite.Assert().NotEmpty(terraform.Output(suite.T(), terraformOptions, "private_route_table_ids"))

	// test IAM roles
	suite.Assert().Equal(fmt.Sprintf("%s-eks-iam-role", clusterName), terraform.Output(suite.T(), terraformOptions, "cluster_iam_role_name"))

	// this is a split(6)[0..2] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPrivateVpcCidrBlocks := "[10.192.0.0/19 10.192.32.0/19 10.192.64.0/19]"
	suite.Assert().Equal(expectedPrivateVpcCidrBlocks, terraform.Output(suite.T(), terraformOptions, "private_vpc_cidr_blocks"))

	// this is a split(6)[3..5] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPublicVpcCidrBlocks := "[10.192.96.0/19 10.192.128.0/19 10.192.160.0/19]"
	suite.Assert().Equal(expectedPublicVpcCidrBlocks, terraform.Output(suite.T(), terraformOptions, "public_vpc_cidr_blocks"))

	sess, err := utils.GetAwsClient()
	suite.Require().NoErrorf(err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)
	iamSvc := iam.NewFromConfig(sess)
	ec2Svc := ec2.NewFromConfig(sess)
	kmsSvc := kms.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	suite.Assert().NoError(err)

	// Wait for the worker nodes to join the cluster
	suite.sugaredLogger.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(suite.expectedNodes))
	suite.Require().NoError(errClusterReady)

	// Verify list of addons installed on the EKS
	expectedEKSAddons := []string{"coredns", "kube-proxy", "vpc-cni", "aws-ebs-csi-driver"}
	inputDescribeAddons := &eks.ListAddonsInput{
		ClusterName: aws.String(clusterName),
	}
	outputEKSAddons, errEKSAddons := eksSvc.ListAddons(context.Background(), inputDescribeAddons)
	suite.Require().NoError(errEKSAddons)

	// perform the diff
	presenceAddonsMap := make(map[string]bool)
	for _, addon := range outputEKSAddons.Addons {
		presenceAddonsMap[addon] = true
	}
	for _, addonName := range expectedEKSAddons {
		suite.Assert().Truef(presenceAddonsMap[addonName], "Addon %s not installed on the EKS cluster", addonName)
	}

	// Verifies EKS roles
	roleNames := []string{
		fmt.Sprintf("%s-cert-manager-role", clusterName),
		fmt.Sprintf("%s-external-dns-role", clusterName),
		fmt.Sprintf("%s-ebs-cs-role", clusterName),
		fmt.Sprintf("%s-eks-iam-role", clusterName),
	}

	for _, roleName := range roleNames {
		input := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err := iamSvc.GetRole(context.Background(), input)
		suite.Assert().NoErrorf(err, "Failed to get IAM EKS role %s", roleName)
	}

	// verifies the VPC

	vpcName := fmt.Sprintf("%s-vpc", clusterName)

	inputVPC := &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{vpcName},
			},
		},
	}

	outputVPC, errVPC := ec2Svc.DescribeVpcs(context.Background(), inputVPC)
	suite.Require().NoError(errVPC)

	suite.Assert().Equal(len(outputVPC.Vpcs), 1)

	// key
	keyDescription := fmt.Sprintf("%s -  EKS Secret Encryption Key", clusterName)
	inputKMS := &kms.ListKeysInput{}
	outputKMSList, errKMSList := kmsSvc.ListKeys(context.Background(), inputKMS)
	suite.Assert().NoError(errKMSList)

	// Check if the key corresponding to the description exists
	keyFound := false
	for _, key := range outputKMSList.Keys {
		keyDetails, errKey := kmsSvc.DescribeKey(context.Background(), &kms.DescribeKeyInput{
			KeyId: key.KeyId,
		})

		if errKey != nil {
			// ignore AccessDenied
			var re *awshttp.ResponseError
			if errors.As(err, &re) {
				if re.HTTPStatusCode() == 400 {
					continue
				}
			}

			suite.Require().NoErrorf(errKey, "Failed to describe key %s", *key.KeyId)
		}

		keyFound = *keyDetails.KeyMetadata.Description == keyDescription
		if keyFound {
			break
		}
	}
	suite.Assert().Truef(keyFound, "Failed to find key %s", keyDescription)
}

func TestDefaultEKSTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(DefaultEKSTestSuite))
}
