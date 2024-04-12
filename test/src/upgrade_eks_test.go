package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/camunda/camunda-tf-eks-module/utils"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
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

type UpgradeEKSTestSuite struct {
	suite.Suite
	logger         *zap.Logger
	sugaredLogger  *zap.SugaredLogger
	clusterName    string
	expectedNodes  int
	kubeConfigPath string
	kubeVersion    string
	tfDataDir      string
	region         string
	varTf          map[string]interface{}
}

func (suite *UpgradeEKSTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.sugaredLogger = suite.logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	suite.clusterName = fmt.Sprintf("cluster-upgrade-%s", clusterSuffix)
	suite.region = utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	suite.expectedNodes = 3
	suite.kubeVersion = "1.28"
	var errAbsPath error
	suite.tfDataDir, errAbsPath = filepath.Abs(fmt.Sprintf("../../test/states/tf-data-%s", suite.clusterName))
	suite.Require().NoError(errAbsPath)
	suite.kubeConfigPath = fmt.Sprintf("%s/kubeconfig-upgrade-eks", suite.tfDataDir)
}

func (suite *UpgradeEKSTestSuite) TearUpTest() {
	// create tf state
	absPath, err := filepath.Abs(suite.tfDataDir)
	suite.Require().NoError(err)
	err = os.MkdirAll(absPath, os.ModePerm)
	suite.Require().NoError(err)
}

func (suite *UpgradeEKSTestSuite) TearDownTest() {
	suite.T().Log("Cleaning up resources...")

	err := os.Remove(suite.kubeConfigPath)
	if err != nil && !os.IsNotExist(err) {
		suite.T().Errorf("Failed to remove kubeConfigPath: %v", err)
	}
}

// TestUpgradeEKS starts from a version of EKS, deploy a simple chart, upgrade the cluster
// and check that everything is working as expected
func (suite *UpgradeEKSTestSuite) TestUpgradeEKS() {
	// create the eks cluster
	suite.varTf = map[string]interface{}{
		"name":                  suite.clusterName,
		"region":                suite.region,
		"np_desired_node_count": suite.expectedNodes,
		"kubernetes_version":    suite.kubeVersion,
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

	suite.sugaredLogger.Infow("Creating EKS cluster...", "extraVars", suite.varTf)

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")

	if cleanClusterAtTheEnd == "true" {
		defer terraform.Destroy(suite.T(), terraformOptions)
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(suite.T(), terraformOptions)
		})
	}

	terraform.InitAndApplyAndIdempotent(suite.T(), terraformOptions)

	// Wait for the worker nodes to join the cluster
	sess, err := utils.GetAwsClient()
	suite.Require().NoErrorf(err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(suite.clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	suite.Assert().NoError(err)

	suite.sugaredLogger.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(suite.expectedNodes))
	suite.Require().NoError(errClusterReady)

	suite.Assert().Equal(suite.kubeVersion, *result.Cluster.Version)

	utils.GenerateKubeConfigFromAWS(suite.T(), suite.region, suite.clusterName, utils.GetAwsProfile(), suite.kubeConfigPath)

	// test suite: deploy a pod and check it is healthy
	namespace := "example"
	kubeCtlOptions := k8s.NewKubectlOptions("", suite.kubeConfigPath, namespace)
	utils.CreateIfNotExistsNamespace(suite.T(), kubeCtlOptions, namespace)

	// deploy the postgres-client Job to test the connection
	k8s.KubectlApply(suite.T(), kubeCtlOptions, "../../modules/fixtures/whoami-deployment.yml")

	k8s.WaitUntilServiceAvailable(suite.T(), kubeCtlOptions, "whoami-service", 60, 1*time.Second)
	// wait to ensure service available
	time.Sleep(30 * time.Second)

	// Now we verify that the service will successfully boot and start serving requests
	localPort1 := 8883

	service := k8s.GetService(suite.T(), kubeCtlOptions, "whoami-service")
	portForwardProc1 := k8s.NewTunnel(kubeCtlOptions, k8s.ResourceTypeService, service.ObjectMeta.Name, localPort1, 80)
	defer portForwardProc1.Close()
	portForwardProc1.ForwardPort(suite.T())

	// wait for the port forward to be ready
	time.Sleep(5 * time.Second)

	http_helper.HttpGetWithRetryWithCustomValidation(
		suite.T(),
		fmt.Sprintf("http://%s", portForwardProc1.Endpoint()),
		nil,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)

	// upgrade the cluster
	var errIncVersion error
	suite.varTf["kubernetes_version"], errIncVersion = utils.IncrementMinorVersionTwoParts(suite.kubeVersion)
	suite.Require().NoError(errIncVersion)

	suite.sugaredLogger.Infow(fmt.Sprintf("Upgrading the EKS cluster to v%s using aws sdk", suite.varTf["kubernetes_version"]), "extraVars", suite.varTf)
	errUpdate := utils.UpgradeEKS(context.Background(), eksSvc, suite.clusterName, suite.varTf["kubernetes_version"].(string))
	suite.Require().NoError(errUpdate)

	suite.sugaredLogger.Infow("Waiting for worker nodes to join the EKS cluster after the upgrade")
	errClusterReady = utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(suite.expectedNodes))
	suite.Require().NoError(errClusterReady)

	// perform update with terraform
	terraformOptions = &terraform.Options{
		TerraformDir: tfDir,
		Upgrade:      false,
		VarFiles:     []string{"../fixtures/fixtures.default.eks.tfvars"},
		Vars:         suite.varTf,
	}

	suite.sugaredLogger.Infow("Reapply terraform after EKS cluster upgrade...", "extraVars", suite.varTf)

	if cleanClusterAtTheEnd == "true" {
		defer terraform.Destroy(suite.T(), terraformOptions)
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(suite.T(), terraformOptions)
		})
	}

	terraform.InitAndApplyAndIdempotent(suite.T(), terraformOptions)

	errClusterReady = utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(suite.expectedNodes))
	suite.Require().NoError(errClusterReady)

	// Check version of the upgraded cluster
	result, err = eksSvc.DescribeCluster(context.Background(), inputEKS)
	suite.Assert().NoError(err)
	suite.Assert().Equal(suite.varTf["kubernetes_version"], *result.Cluster.Version)

	// check everything works as expected
	k8s.WaitUntilServiceAvailable(suite.T(), kubeCtlOptions, "whoami-service", 60, 1*time.Second)
	// wait to ensure service available
	time.Sleep(30 * time.Second)

	// Forward port again
	localPort2 := 8887
	service = k8s.GetService(suite.T(), kubeCtlOptions, "whoami-service")
	portForwardProc2 := k8s.NewTunnel(kubeCtlOptions, k8s.ResourceTypeService, service.ObjectMeta.Name, localPort2, 80)
	defer portForwardProc2.Close()
	portForwardProc2.ForwardPort(suite.T())

	// Wait for port forward to be ready
	time.Sleep(5 * time.Second)

	http_helper.HttpGetWithRetryWithCustomValidation(
		suite.T(),
		fmt.Sprintf("http://%s", portForwardProc2.Endpoint()),
		nil,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}

func TestUpgradeEKSTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UpgradeEKSTestSuite))
}
