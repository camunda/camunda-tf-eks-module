package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	"github.com/camunda/camunda-tf-eks-module/utils"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type CustomEKSOpenSearchTestSuite struct {
	suite.Suite
	logger          *zap.Logger
	sugaredLogger   *zap.SugaredLogger
	clusterName     string
	expectedNodes   int
	kubeConfigPath  string
	region          string
	bucketRegion    string
	tfDataDir       string
	tfBinaryName    string
	varTf           map[string]interface{}
	tfStateS3Bucket string
}

func (suite *CustomEKSOpenSearchTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.sugaredLogger = suite.logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	suite.clusterName = fmt.Sprintf("cluster-opensearch-%s", clusterSuffix)
	suite.region = utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	suite.bucketRegion = utils.GetEnv("TF_STATE_BUCKET_REGION", suite.region)
	suite.tfBinaryName = utils.GetEnv("TESTS_TF_BINARY_NAME", "terraform")
	suite.sugaredLogger.Infow("Terraform binary for the suite", "binary", suite.tfBinaryName)

	suite.expectedNodes = 1
	var errAbsPath error
	suite.tfStateS3Bucket = utils.GetEnv("TF_STATE_BUCKET", fmt.Sprintf("tests-eks-tf-state-%s", suite.bucketRegion))
	suite.tfDataDir, errAbsPath = filepath.Abs(fmt.Sprintf("../../test/states/tf-data-%s", suite.clusterName))
	suite.Require().NoError(errAbsPath)
	suite.kubeConfigPath = fmt.Sprintf("%s/kubeconfig-opensearch-eks", suite.tfDataDir)
}

func (suite *CustomEKSOpenSearchTestSuite) TearUpTest() {
	// create tf state
	absPath, err := filepath.Abs(suite.tfDataDir)
	suite.Require().NoError(err)
	err = os.MkdirAll(absPath, os.ModePerm)
	suite.Require().NoError(err)
}

func (suite *CustomEKSOpenSearchTestSuite) TearDownTest() {
	suite.T().Log("Cleaning up resources...")

	err := os.Remove(suite.kubeConfigPath)
	if err != nil && !os.IsNotExist(err) {
		suite.T().Errorf("Failed to remove kubeConfigPath: %v", err)
	}
}

// TestCustomEKSAndOpenSearch spawns a custom EKS cluster with custom parameters, and spawns a
// a curl pod that will try to reach the OpenSearch cluster
// TODO: implement IRSA connection in the pod https://github.com/opensearch-project/logstash-output-opensearch/issues/96
func (suite *CustomEKSOpenSearchTestSuite) TestCustomEKSAndOpenSearch() {
	suite.varTf = map[string]interface{}{
		"name":                  suite.clusterName,
		"region":                suite.region,
		"np_desired_node_count": suite.expectedNodes,
	}

	suite.sugaredLogger.Infow("Creating EKS cluster...", "extraVars", suite.varTf)

	tfModuleEKS := "eks-cluster/"
	fullDirEKS := fmt.Sprintf("%s%s", suite.tfDataDir, tfModuleEKS)
	errTfDirEKS := os.MkdirAll(fullDirEKS, os.ModePerm)
	suite.Require().NoError(errTfDirEKS)
	tfDir := test_structure.CopyTerraformFolderToDest(suite.T(), "../../modules/", tfModuleEKS, fullDirEKS)

	errLinkBackend := os.Link("../../modules/fixtures/backend.tf", filepath.Join(tfDir, "backend.tf"))
	suite.Require().NoError(errLinkBackend)

	terraformOptions := &terraform.Options{
		TerraformBinary: suite.tfBinaryName,
		TerraformDir:    tfDir,
		Upgrade:         false,
		VarFiles:        []string{"../fixtures/fixtures.default.eks.tfvars"},
		Vars:            suite.varTf,
		BackendConfig: map[string]interface{}{
			"bucket": suite.tfStateS3Bucket,
			"key":    fmt.Sprintf("terraform/%s/TestCustomEKSOpenSearchTestSuite/%sterraform.tfstate", suite.clusterName, tfModuleEKS),
			"region": suite.bucketRegion,
		},
	}

	// configure bucket backend
	sessBackend, err := utils.GetAwsClientF(utils.GetAwsProfile(), suite.bucketRegion)
	suite.Require().NoErrorf(err, "Failed to get aws client")
	err = utils.CreateS3BucketIfNotExists(sessBackend, suite.tfStateS3Bucket, utils.TF_BUCKET_DESCRIPTION, suite.bucketRegion)
	suite.Require().NoErrorf(err, "Failed to create s3 state bucket")

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")
	if cleanClusterAtTheEnd == "true" {
		defer utils.DeferCleanup(suite.T(), suite.bucketRegion, terraformOptions)
	}

	terraform.InitAndApply(suite.T(), terraformOptions)

	sess, err := utils.GetAwsClient()
	suite.Require().NoErrorf(err, "Failed to get aws client")

	eksSvc := eks.NewFromConfig(sess)
	opensearchSvc := opensearch.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(suite.clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	suite.Assert().NoError(err)

	// Spawn OpenSearch within the EKS VPC/subnet
	publicBlocks := strings.Fields(strings.Trim(terraform.Output(suite.T(), terraformOptions, "public_vpc_cidr_blocks"), "[]"))
	privateBlocks := strings.Fields(strings.Trim(terraform.Output(suite.T(), terraformOptions, "private_vpc_cidr_blocks"), "[]"))

	opensearchDomainName := fmt.Sprintf("opensearch-%s", suite.clusterName)
	opensearchMasterUserName := "opensearch-admin"
	opensearchMasterUserPassword := "password"

	varsConfigOpenSearch := map[string]interface{}{
		"domain_name":                            opensearchDomainName,
		"advanced_security_master_user_name":     opensearchMasterUserName,
		"advanced_security_master_user_password": opensearchMasterUserPassword,
		"subnet_ids":                             result.Cluster.ResourcesVpcConfig.SubnetIds,
		"vpc_id":                                 *result.Cluster.ResourcesVpcConfig.VpcId,
		"availability_zones":                     []string{fmt.Sprintf("%sa", suite.region), fmt.Sprintf("%sb", suite.region), fmt.Sprintf("%sc", suite.region)},
		"cidr_blocks":                            append(publicBlocks, privateBlocks...),
	}

	tfModuleOpenSearch := "opensearch/"
	fullDirOpenSearch := fmt.Sprintf("%s/%s", suite.tfDataDir, tfModuleOpenSearch)
	errTfDirOpenSearch := os.MkdirAll(fullDirOpenSearch, os.ModePerm)
	suite.Require().NoError(errTfDirOpenSearch)

	tfDirOpenSearch := test_structure.CopyTerraformFolderToDest(suite.T(), "../../modules/", tfModuleOpenSearch, fullDirOpenSearch)

	errLinkBackend = os.Link("../../modules/fixtures/backend.tf", filepath.Join(tfDirOpenSearch, "backend.tf"))
	suite.Require().NoError(errLinkBackend)

	terraformOptionsOpenSearch := &terraform.Options{
		TerraformBinary: suite.tfBinaryName,
		TerraformDir:    tfDirOpenSearch,
		Upgrade:         false,
		VarFiles:        []string{"../fixtures/fixtures.default.opensearch.tfvars"},
		Vars:            varsConfigOpenSearch,
		BackendConfig: map[string]interface{}{
			"bucket": suite.tfStateS3Bucket,
			"key":    fmt.Sprintf("terraform/%s/TestCustomEKSOpenSearchTestSuite/%sterraform.tfstate", suite.clusterName, tfModuleOpenSearch),
			"region": suite.bucketRegion,
		},
	}

	if cleanClusterAtTheEnd == "true" {
		defer utils.DeferCleanup(suite.T(), suite.bucketRegion, terraformOptionsOpenSearch)
	}

	terraform.InitAndApply(suite.T(), terraformOptionsOpenSearch)
	opensearchEndpoint := terraform.Output(suite.T(), terraformOptionsOpenSearch, "opensearch_domain_endpoint")
	suite.Assert().NotEmpty(opensearchEndpoint)

	// Test the OpenSearch connection and perform additional tests as needed

	// TODO

	// Retrieve OpenSearch information
	describeDomainInput := &opensearch.DescribeDomainInput{
		DomainName: aws.String(varsConfigOpenSearch["domain_name"].(string)),
	}
	describeDomainOutput, err := opensearchSvc.DescribeDomain(context.Background(), describeDomainInput)
	suite.Require().NoError(err)

	// Perform assertions on the OpenSearch domain configuration

	// TODO
}

func TestCustomEKSOpenSearchTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CustomEKSOpenSearchTestSuite))
}
