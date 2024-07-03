package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/camunda/camunda-tf-eks-module/utils"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type CustomEKSRDSTestSuite struct {
	suite.Suite
	logger         *zap.Logger
	sugaredLogger  *zap.SugaredLogger
	clusterName    string
	expectedNodes  int
	kubeConfigPath string
	region         string
	tfDataDir      string
	tfBinaryName   string
	varTf          map[string]interface{}
}

func (suite *CustomEKSRDSTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.sugaredLogger = suite.logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	suite.clusterName = fmt.Sprintf("cluster-rds-%s", clusterSuffix)
	suite.region = utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	suite.tfBinaryName = utils.GetEnv("TESTS_TF_BINARY_NAME", "terraform")
	suite.sugaredLogger.Infow("Terraform binary for the suite", "binary", suite.tfBinaryName)

	suite.expectedNodes = 1
	var errAbsPath error
	suite.tfDataDir, errAbsPath = filepath.Abs(fmt.Sprintf("../../test/states/tf-data-%s", suite.clusterName))
	suite.Require().NoError(errAbsPath)
	suite.kubeConfigPath = fmt.Sprintf("%s/kubeconfig-rds-eks", suite.tfDataDir)
}

func (suite *CustomEKSRDSTestSuite) TearUpTest() {
	// create tf state
	absPath, err := filepath.Abs(suite.tfDataDir)
	suite.Require().NoError(err)
	err = os.MkdirAll(absPath, os.ModePerm)
	suite.Require().NoError(err)
}

func (suite *CustomEKSRDSTestSuite) TearDownTest() {
	suite.T().Log("Cleaning up resources...")

	err := os.Remove(suite.kubeConfigPath)
	if err != nil && !os.IsNotExist(err) {
		suite.T().Errorf("Failed to remove kubeConfigPath: %v", err)
	}
}

// TestCustomEKSAndRDS spawns a custom EKS cluster with custom parameters, and spawns a
// pg client pod that will test connection to AuroraDB
func (suite *CustomEKSRDSTestSuite) TestCustomEKSAndRDS() {
	suite.varTf = map[string]interface{}{
		"name":                  suite.clusterName,
		"region":                suite.region,
		"np_desired_node_count": suite.expectedNodes,
	}

	suite.sugaredLogger.Infow("Creating EKS cluster...", "extraVars", suite.varTf)

	fullDirEKS := fmt.Sprintf("%seks-cluster/", suite.tfDataDir)
	errTfDirEKS := os.MkdirAll(fullDirEKS, os.ModePerm)
	suite.Require().NoError(errTfDirEKS)
	tfDir := test_structure.CopyTerraformFolderToDest(suite.T(), "../../modules/", "eks-cluster/", fullDirEKS)

	terraformOptions := &terraform.Options{
		TerraformBinary: suite.tfBinaryName,
		TerraformDir:    tfDir,
		Upgrade:         false,
		VarFiles:        []string{"../fixtures/fixtures.default.eks.tfvars"},
		Vars:            suite.varTf,
	}

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")

	if cleanClusterAtTheEnd == "true" {
		defer terraform.Destroy(suite.T(), terraformOptions)
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(suite.T(), terraformOptions)
		})
	}

	// since v20, we can't use InitAndApplyAndIdempotent due to labels being added
	terraform.InitAndApplyAndIdempotent(suite.T(), terraformOptions)

	// Wait for the worker nodes to join the cluster
	sess, err := utils.GetAwsClient()
	suite.Require().NoErrorf(err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)
	rdsSvc := rds.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(suite.clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	suite.Assert().NoError(err)

	suite.sugaredLogger.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(suite.expectedNodes))
	suite.Require().NoError(errClusterReady)

	// Spawn RDS within the EKS VPC/subnet
	publicBlocks := strings.Fields(strings.Trim(terraform.Output(suite.T(), terraformOptions, "public_vpc_cidr_blocks"), "[]"))
	privateBlocks := strings.Fields(strings.Trim(terraform.Output(suite.T(), terraformOptions, "private_vpc_cidr_blocks"), "[]"))

	auroraUsername := "myuser"
	auroraPassword := "mypassword123secure"
	auroraDatabase := "camunda"

	varsConfigAurora := map[string]interface{}{
		"username":              auroraUsername,
		"password":              auroraPassword,
		"default_database_name": auroraDatabase,
		"cluster_name":          fmt.Sprintf("postgres-%s", suite.clusterName),
		"subnet_ids":            result.Cluster.ResourcesVpcConfig.SubnetIds,
		"vpc_id":                *result.Cluster.ResourcesVpcConfig.VpcId,
		"availability_zones":    []string{fmt.Sprintf("%sa", suite.region), fmt.Sprintf("%sb", suite.region), fmt.Sprintf("%sc", suite.region)},
		"cidr_blocks":           append(publicBlocks, privateBlocks...),
		"iam_auth_enabled":      true,
	}

	fullDirAurora := fmt.Sprintf("%s/aurora/", suite.tfDataDir)
	errTfDirAurora := os.MkdirAll(fullDirAurora, os.ModePerm)
	suite.Require().NoError(errTfDirAurora)

	tfDirAurora := test_structure.CopyTerraformFolderToDest(suite.T(), "../../modules/", "aurora/", fullDirAurora)

	terraformOptionsRDS := &terraform.Options{
		TerraformBinary: suite.tfBinaryName,
		TerraformDir:    tfDirAurora,
		Upgrade:         false,
		VarFiles:        []string{"../fixtures/fixtures.default.aurora.tfvars"},
		Vars:            varsConfigAurora,
	}

	if cleanClusterAtTheEnd == "true" {
		defer terraform.Destroy(suite.T(), terraformOptionsRDS)
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(suite.T(), terraformOptionsRDS)
		})
	}

	terraform.InitAndApplyAndIdempotent(suite.T(), terraformOptionsRDS)
	auroraEndpoint := terraform.Output(suite.T(), terraformOptionsRDS, "aurora_endpoint")
	suite.Assert().NotEmpty(auroraEndpoint)

	// Test of the RDS connection is performed by launching a pod on the cluster and test the pg connection
	kubeClient, err := utils.NewKubeClientSet(result.Cluster)
	suite.Require().NoError(err)

	utils.GenerateKubeConfigFromAWS(suite.T(), suite.region, suite.clusterName, utils.GetAwsProfile(), suite.kubeConfigPath)

	namespace := "postgres-client"
	pgKubeCtlOptions := k8s.NewKubectlOptions("", suite.kubeConfigPath, namespace)
	utils.CreateIfNotExistsNamespace(suite.T(), pgKubeCtlOptions, namespace)

	// deploy the postgres-client ConfigMap
	configMapPostgres := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aurora-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"aurora_endpoint":      auroraEndpoint,
			"aurora_username":      auroraUsername,
			"aurora_username_irsa": fmt.Sprintf("%s-irsa", auroraUsername),
			"aurora_port":          "5432",
			"aws_region":           suite.region,
			"aurora_db_name":       auroraDatabase,
		},
	}

	err = kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configMapPostgres.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		suite.Require().NoError(err)
	}
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMapPostgres, metav1.CreateOptions{})
	k8s.WaitUntilConfigMapAvailable(suite.T(), pgKubeCtlOptions, configMapPostgres.Name, 6, 10*time.Second)

	// create the secret for aurora pg password
	secretPostgres := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aurora-secret",
			Namespace: namespace,
		},
		StringData: map[string]string{
			"aurora_password": auroraPassword,
		},
	}
	err = kubeClient.CoreV1().Secrets(namespace).Delete(context.Background(), configMapPostgres.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		suite.Require().NoError(err)
	}
	_, err = kubeClient.CoreV1().Secrets(namespace).Create(context.Background(), secretPostgres, metav1.CreateOptions{})
	k8s.WaitUntilSecretAvailable(suite.T(), pgKubeCtlOptions, secretPostgres.Name, 6, 10*time.Second)

	// add the scripts as a ConfigMap
	scriptPath := "../../modules/fixtures/scripts/create_aurora_pg_db.sh"
	scriptContent, err := os.ReadFile(scriptPath)
	suite.Require().NoError(err)

	configMapScript := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgres-scripts",
			Namespace: namespace,
		},
		Data: map[string]string{
			"create_aurora_pg_db.sh": string(scriptContent),
		},
	}

	err = kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configMapScript.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		suite.Require().NoError(err)
	}
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMapScript, metav1.CreateOptions{})
	k8s.WaitUntilConfigMapAvailable(suite.T(), pgKubeCtlOptions, configMapScript.Name, 6, 10*time.Second)

	// cleanup existing jobs
	jobListOptions := metav1.ListOptions{LabelSelector: "app=postgres-client"}
	existingJobs := k8s.ListJobs(suite.T(), pgKubeCtlOptions, jobListOptions)
	for _, job := range existingJobs {
		err := kubeClient.BatchV1().Jobs(namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{})
		suite.Assert().NoError(err)
	}

	// deploy the postgres-client Job to test the connection
	k8s.KubectlApply(suite.T(), pgKubeCtlOptions, "../../modules/fixtures/postgres-client.yml")
	errJob := utils.WaitForJobCompletion(kubeClient, namespace, "postgres-client", 5*time.Minute, jobListOptions)
	suite.Require().NoError(errJob)

	// TODO: test IRSA apply https://kubedemy.io/aws-eks-part-13-setup-iam-roles-for-service-accounts-irsa to setup iam

	// Retrieve RDS information
	describeDBClusterInput := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(varsConfigAurora["cluster_name"].(string)),
	}
	describeDBClusterOutput, err := rdsSvc.DescribeDBClusters(context.Background(), describeDBClusterInput)
	suite.Require().NoError(err)

	expectedRDSAZ := []string{fmt.Sprintf("%sa", suite.region), fmt.Sprintf("%sb", suite.region), fmt.Sprintf("%sc", suite.region)}
	suite.Assert().Equal(varsConfigAurora["iam_auth_enabled"].(bool), *describeDBClusterOutput.DBClusters[0].IAMDatabaseAuthenticationEnabled)
	suite.Assert().Equal(varsConfigAurora["username"].(string), *describeDBClusterOutput.DBClusters[0].MasterUsername)
	suite.Assert().Equal(auroraDatabase, *describeDBClusterOutput.DBClusters[0].DatabaseName)
	suite.Assert().Equal(int32(5432), *describeDBClusterOutput.DBClusters[0].Port)
	suite.Assert().Equal("15.4", *describeDBClusterOutput.DBClusters[0].EngineVersion)
	suite.Assert().ElementsMatch(expectedRDSAZ, describeDBClusterOutput.DBClusters[0].AvailabilityZones)
	suite.Assert().Equal(varsConfigAurora["cluster_name"].(string), *describeDBClusterOutput.DBClusters[0].DBClusterIdentifier)

	// Some of the tests are performed on the first instance of the cluster
	describeDBInstanceInput := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: describeDBClusterOutput.DBClusters[0].DBClusterMembers[0].DBInstanceIdentifier,
	}
	describeDBInstanceOutput, err := rdsSvc.DescribeDBInstances(context.Background(), describeDBInstanceInput)
	suite.Require().NoError(err)

	suite.Assert().Equal("db.t3.medium", *describeDBInstanceOutput.DBInstances[0].DBInstanceClass)
	suite.Assert().Equal(true, *describeDBInstanceOutput.DBInstances[0].AutoMinorVersionUpgrade)
	suite.Assert().Equal("aurora-postgresql", *describeDBInstanceOutput.DBInstances[0].Engine)
	suite.Assert().Equal("rds-ca-2019", *describeDBInstanceOutput.DBInstances[0].CertificateDetails.CAIdentifier)
	suite.Assert().Equal(varsConfigAurora["vpc_id"].(string), *describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.VpcId)
	suite.Assert().Contains(*describeDBInstanceOutput.DBInstances[0].AvailabilityZone, suite.region)

	// construct the subnet ids
	actualSubnetIds := make([]string, len(describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.Subnets))
	for id, subnet := range describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.Subnets {
		actualSubnetIds[id] = *subnet.SubnetIdentifier
	}
	suite.Assert().ElementsMatch(varsConfigAurora["subnet_ids"].([]string), actualSubnetIds)

	// EKS test that custom cluster parameters are applied as expected

	// count nb of nodes
	nodes, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	suite.Require().NoError(err)
	suite.Assert().Equal(suite.expectedNodes, len(nodes.Items))

	// verifies for each node, the flavor and the region
	expectedInstanceType := "t2.medium"
	for _, node := range nodes.Items {
		regionNode, _ := node.Labels["failure-domain.beta.kubernetes.io/region"]
		instanceType, _ := node.Labels["node.kubernetes.io/instance-type"]
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				suite.Assert().Equal(suite.region, regionNode)
				suite.Assert().Equal(expectedInstanceType, instanceType)
			}
		}
	}
}

func TestCustomEKSRDSTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CustomEKSRDSTestSuite))
}
