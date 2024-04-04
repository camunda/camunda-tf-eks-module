package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/camunda/camunda-tf-eks-module/utils"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"strings"
	"testing"
	"time"
)

// TestDefaultEKS spawns an EKS cluster with the default parameters and checks the parameters
func TestDefaultEKS(t *testing.T) {
	// log
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	clusterName := fmt.Sprintf("cluster-%s", clusterSuffix)
	region := utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	sugar.Infow("Creating EKS cluster...")
	expectedCapacity := 4

	varsConfig := map[string]interface{}{
		"name":                  clusterName,
		"region":                region,
		"np_desired_node_count": expectedCapacity,
	}

	terraformOptions := SpawnEKS(t, sugar, varsConfig)

	// test suite
	baseChecksEKS(t, sugar, terraformOptions, uint64(expectedCapacity))

	TearsDown(t, sugar)
}

// TestCustomEKSAndRDS spawns a custom EKS cluster with custom parameters, and spawns a
// pg client pod that will test connection to AuroraDB
func TestCustomEKSAndRDS(t *testing.T) {
	// log
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	clusterName := fmt.Sprintf("cluster-rds-%s", clusterSuffix)
	region := utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	sugar.Infow("Creating EKS cluster...")
	expectedCapacity := 3

	varsConfigEKS := map[string]interface{}{
		"name":                  clusterName,
		"region":                region,
		"np_desired_node_count": expectedCapacity,
	}

	terraformOptions := SpawnEKS(t, sugar, varsConfigEKS)

	// Wait for the worker nodes to join the cluster
	sess, err := utils.GetAwsClient()
	require.NoErrorf(t, err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)
	rdsSvc := rds.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	assert.NoError(t, err)

	sugar.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(expectedCapacity))
	require.NoError(t, errClusterReady)

	// Spawn RDS within the EKS VPC/subnet
	publicBlocks := strings.Fields(strings.Trim(terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"), "[]"))
	privateBlocks := strings.Fields(strings.Trim(terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"), "[]"))

	auroraUsername := "myuser"
	auroraPassword := "mypassword123secure"
	auroraDatabase := "camunda"

	varsConfigAurora := map[string]interface{}{
		"username":              auroraUsername,
		"password":              auroraPassword,
		"default_database_name": auroraDatabase,
		"cluster_name":          fmt.Sprintf("postgres-%s", clusterSuffix),
		"subnet_ids":            result.Cluster.ResourcesVpcConfig.SubnetIds,
		"vpc_id":                *result.Cluster.ResourcesVpcConfig.VpcId,
		"cidr_blocks":           append(publicBlocks, privateBlocks...),
		"iam_auth_enabled":      true,
	}

	terraformOptionsRDS := SpawnAurora(t, sugar, varsConfigAurora)
	auroraEndpoint := terraform.Output(t, terraformOptionsRDS, "aurora_endpoint")
	assert.NotEmpty(t, auroraEndpoint)

	// Test of the RDS connection is performed by launching a pod on the cluster and test the pg connection
	kubeClient, err := utils.NewKubeClientSet(result.Cluster)
	require.NoError(t, err)

	kubeConfigPath := "kubeconfig-eks-rds"
	utils.GenerateKubeConfigFromAWS(t, region, clusterName, utils.GetAwsProfile(), kubeConfigPath)

	namespace := "postgres-client"
	pgKubeCtlOptions := k8s.NewKubectlOptions("", kubeConfigPath, namespace)
	utils.CreateIfNotExistsNamespace(t, pgKubeCtlOptions, namespace)

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
			"aws_region":           region,
			"aurora_db_name":       auroraDatabase,
		},
	}

	err = kubeClient.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configMapPostgres.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		require.NoError(t, err)
	}
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMapPostgres, metav1.CreateOptions{})
	k8s.WaitUntilConfigMapAvailable(t, pgKubeCtlOptions, configMapPostgres.Name, 6, 10*time.Second)

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
		require.NoError(t, err)
	}
	_, err = kubeClient.CoreV1().Secrets(namespace).Create(context.Background(), secretPostgres, metav1.CreateOptions{})
	k8s.WaitUntilSecretAvailable(t, pgKubeCtlOptions, secretPostgres.Name, 6, 10*time.Second)

	// add the scripts as a ConfigMap
	scriptPath := "../../test/src/fixtures/scripts/create_aurora_pg_db.sh"
	scriptContent, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

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
		require.NoError(t, err)
	}
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMapScript, metav1.CreateOptions{})
	k8s.WaitUntilConfigMapAvailable(t, pgKubeCtlOptions, configMapScript.Name, 6, 10*time.Second)

	// cleanup existing jobs
	jobListOptions := metav1.ListOptions{LabelSelector: "app=postgres-client"}
	existingJobs := k8s.ListJobs(t, pgKubeCtlOptions, jobListOptions)
	for _, job := range existingJobs {
		err := kubeClient.BatchV1().Jobs(namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{})
		assert.NoError(t, err)
	}

	// deploy the postgres-client Job to test the connection
	k8s.KubectlApply(t, pgKubeCtlOptions, "../../test/src/fixtures/postgres-client.yml")
	errJob := utils.WaitForJobCompletion(kubeClient, namespace, "postgres-client", 5*time.Minute, jobListOptions)
	require.NoError(t, errJob)

	// TODO: test IRSA apply https://kubedemy.io/aws-eks-part-13-setup-iam-roles-for-service-accounts-irsa to setup iam

	// Retrieve RDS information
	describeDBClusterInput := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(varsConfigAurora["cluster_name"].(string)),
	}
	describeDBClusterOutput, err := rdsSvc.DescribeDBClusters(context.Background(), describeDBClusterInput)
	require.NoError(t, err)

	expectedRDSAZ := []string{fmt.Sprintf("%sa", region), fmt.Sprintf("%sb", region), fmt.Sprintf("%sc", region)}
	assert.Equal(t, varsConfigAurora["iam_auth_enabled"].(bool), *describeDBClusterOutput.DBClusters[0].IAMDatabaseAuthenticationEnabled)
	assert.Equal(t, varsConfigAurora["username"].(string), *describeDBClusterOutput.DBClusters[0].MasterUsername)
	assert.Equal(t, auroraDatabase, *describeDBClusterOutput.DBClusters[0].DatabaseName)
	assert.Equal(t, int32(5432), *describeDBClusterOutput.DBClusters[0].Port)
	assert.Equal(t, "15.4", *describeDBClusterOutput.DBClusters[0].EngineVersion)
	assert.ElementsMatch(t, expectedRDSAZ, describeDBClusterOutput.DBClusters[0].AvailabilityZones)
	assert.Equal(t, varsConfigAurora["cluster_name"].(string), *describeDBClusterOutput.DBClusters[0].DBClusterIdentifier)

	// Some of the tests are performed on the first instance of the cluster
	describeDBInstanceInput := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: describeDBClusterOutput.DBClusters[0].DBClusterMembers[0].DBInstanceIdentifier,
	}
	describeDBInstanceOutput, err := rdsSvc.DescribeDBInstances(context.Background(), describeDBInstanceInput)
	require.NoError(t, err)

	assert.Equal(t, "db.t3.medium", *describeDBInstanceOutput.DBInstances[0].DBInstanceClass)
	assert.Equal(t, true, *describeDBInstanceOutput.DBInstances[0].AutoMinorVersionUpgrade)
	assert.Equal(t, "aurora-postgresql", *describeDBInstanceOutput.DBInstances[0].Engine)
	assert.Equal(t, "rds-ca-2019", *describeDBInstanceOutput.DBInstances[0].CertificateDetails.CAIdentifier)
	assert.Equal(t, varsConfigAurora["vpc_id"].(string), *describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.VpcId)
	assert.Contains(t, *describeDBInstanceOutput.DBInstances[0].AvailabilityZone, region)

	// construct the subnet ids
	actualSubnetIds := make([]string, len(describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.Subnets))
	for id, subnet := range describeDBInstanceOutput.DBInstances[0].DBSubnetGroup.Subnets {
		actualSubnetIds[id] = *subnet.SubnetIdentifier
	}
	assert.ElementsMatch(t, varsConfigAurora["subnet_ids"].([]string), actualSubnetIds)

	// EKS test that custom cluster parameters are applied as expected

	// count nb of nodes
	nodes, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	assert.Equal(t, expectedCapacity, len(nodes.Items))

	// verifies for each node, the flavor and the region
	expectedInstanceType := "t2.medium"
	for _, node := range nodes.Items {
		regionNode, _ := node.Labels["failure-domain.beta.kubernetes.io/region"]
		instanceType, _ := node.Labels["node.kubernetes.io/instance-type"]
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				assert.Equal(t, region, regionNode)
				assert.Equal(t, expectedInstanceType, instanceType)
			}
		}
	}
}

// TestUpgradeEKS starts from a version of EKS, deploy a simple chart, upgrade the cluster
// and check that everything is working as expected
func TestUpgradeEKS(t *testing.T) {
	// log
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	clusterName := fmt.Sprintf("cluster-upgrade-%s", clusterSuffix)
	region := utils.GetEnv("TESTS_CLUSTER_REGION", "eu-central-1")
	sugar.Infow("Creating EKS cluster...")
	expectedCapacity := 3

	varsConfig := map[string]interface{}{
		"name":                  clusterName,
		"region":                region,
		"np_desired_node_count": expectedCapacity,
		"kubernetes_version":    "1.27",
	}

	SpawnEKS(t, sugar, varsConfig)

	// Wait for the worker nodes to join the cluster
	sess, err := utils.GetAwsClient()
	require.NoErrorf(t, err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	assert.NoError(t, err)

	sugar.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(expectedCapacity))
	require.NoError(t, errClusterReady)

	assert.Equal(t, varsConfig["kubernetes_version"], *result.Cluster.Version)

	kubeConfigPath := "kubeconfig-upgrade-eks"
	utils.GenerateKubeConfigFromAWS(t, region, clusterName, utils.GetAwsProfile(), kubeConfigPath)

	// test suite: deploy a pod and check it is healthy
	namespace := "example"
	kubeCtlOptions := k8s.NewKubectlOptions("", kubeConfigPath, namespace)
	utils.CreateIfNotExistsNamespace(t, kubeCtlOptions, namespace)

	// deploy the postgres-client Job to test the connection
	k8s.KubectlApply(t, kubeCtlOptions, "../../test/src/fixtures/whoami-deployment.yml")

	k8s.WaitUntilServiceAvailable(t, kubeCtlOptions, "whoami-service", 10, 1*time.Second)

	// Now we verify that the service will successfully boot and start serving requests
	localPort1 := 8883

	service := k8s.GetService(t, kubeCtlOptions, "whoami-service")
	portForwardProc1 := k8s.NewTunnel(kubeCtlOptions, k8s.ResourceTypeService, service.ObjectMeta.Name, localPort1, 80)
	defer portForwardProc1.Close()
	portForwardProc1.ForwardPort(t)

	// wait for the port forward to be ready
	time.Sleep(5 * time.Second)

	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", portForwardProc1.Endpoint()),
		nil,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)

	// upgrade the cluster
	varsConfig["kubernetes_version"] = "1.28"
	SpawnEKS(t, sugar, varsConfig)

	sugar.Infow("Waiting for worker nodes to join the EKS cluster after the upgrade")
	errClusterReady = utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, uint64(expectedCapacity))
	require.NoError(t, errClusterReady)

	// check version of the upgraded cluster
	result, err = eksSvc.DescribeCluster(context.Background(), inputEKS)
	assert.NoError(t, err)
	assert.Equal(t, varsConfig["kubernetes_version"], *result.Cluster.Version)

	// check everything works as expected
	k8s.WaitUntilServiceAvailable(t, kubeCtlOptions, "whoami-service", 10, 1*time.Second)

	// Now we verify that the service will successfully boot and start serving requests
	localPort2 := 8887

	service = k8s.GetService(t, kubeCtlOptions, "whoami-service")
	portForwardProc2 := k8s.NewTunnel(kubeCtlOptions, k8s.ResourceTypeService, service.ObjectMeta.Name, localPort2, 80)
	defer portForwardProc2.Close()
	portForwardProc2.ForwardPort(t)

	// wait for the port forward to be ready
	time.Sleep(5 * time.Second)

	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", portForwardProc2.Endpoint()),
		nil,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)

	TearsDown(t, sugar)
}

// SpawnEKS spawns a new EKS Cluster from a default fixture file
func SpawnEKS(t *testing.T, sugar *zap.SugaredLogger, varsConfig map[string]interface{}) *terraform.Options {
	sugar.Infow("TF vars", "vars", varsConfig)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/eks-cluster",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.default.eks.tfvars"},
		Vars:     varsConfig,
	}

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")

	if cleanClusterAtTheEnd == "true" {
		// At the end of the test, run `terraform destroy` to clean up any resources that were created
		defer terraform.Destroy(t, terraformOptions)

		// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(t, terraformOptions)
		})
	}

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	// then it will re-run apply to make sure that out tf is idempotent
	terraform.InitAndApplyAndIdempotent(t, terraformOptions)
	return terraformOptions
}

// SpawnAurora spawns a new Aurora RDS from a default fixture file
func SpawnAurora(t *testing.T, sugar *zap.SugaredLogger, varsConfig map[string]interface{}) *terraform.Options {
	sugar.Infow("TF vars", "vars", varsConfig)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/aurora",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.default.aurora.tfvars"},
		Vars:     varsConfig,
	}

	cleanClusterAtTheEnd := utils.GetEnv("CLEAN_CLUSTER_AT_THE_END", "true")

	if cleanClusterAtTheEnd == "true" {
		// At the end of the test, run `terraform destroy` to clean up any resources that were created
		defer terraform.Destroy(t, terraformOptions)

		// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(t, terraformOptions)
		})
	}

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	// then it will re-run apply to make sure that out tf is idempotent
	terraform.InitAndApplyAndIdempotent(t, terraformOptions)
	return terraformOptions
}

func TearsDown(t *testing.T, sugar *zap.SugaredLogger) {
	sugar.Infow("All tests completed")
}

// baseChecksEKS checks the defaults of an EKS cluster
func baseChecksEKS(t *testing.T, sugar *zap.SugaredLogger, terraformOptions *terraform.Options, expectedNodesCount uint64) {
	clusterName := terraformOptions.Vars["name"].(string)
	sugar.Infow("Testing status of the EKS cluster", "clusterName", clusterName)

	// Do some basic not empty tests on outputs
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_endpoint"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_security_group_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_primary_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_iam_role_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "ebs_cs_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "external_dns_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "vpc_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_subnet_ids"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "default_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "vpc_main_route_table_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_route_table_ids"))

	// test IAM roles
	assert.Equal(t, fmt.Sprintf("%s-eks-iam-role", clusterName), terraform.Output(t, terraformOptions, "cluster_iam_role_name"))

	// this is a split(6)[0..2] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPrivateVpcCidrBlocks := "[10.192.0.0/19 10.192.32.0/19 10.192.64.0/19]"
	assert.Equal(t, expectedPrivateVpcCidrBlocks, terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"))

	// this is a split(6)[3..5] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPublicVpcCidrBlocks := "[10.192.96.0/19 10.192.128.0/19 10.192.160.0/19]"
	assert.Equal(t, expectedPublicVpcCidrBlocks, terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"))

	sess, err := utils.GetAwsClient()
	require.NoErrorf(t, err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)
	iamSvc := iam.NewFromConfig(sess)
	ec2Svc := ec2.NewFromConfig(sess)
	kmsSvc := kms.NewFromConfig(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.Background(), inputEKS)
	assert.NoError(t, err)

	// Wait for the worker nodes to join the cluster
	sugar.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilKubeClusterIsReady(result.Cluster, 5*time.Minute, expectedNodesCount)
	require.NoError(t, errClusterReady)

	// Verify list of addons installed on the EKS
	expectedEKSAddons := []string{"coredns", "kube-proxy", "vpc-cni", "aws-ebs-csi-driver"}
	inputDescribeAddons := &eks.ListAddonsInput{
		ClusterName: aws.String(clusterName),
	}
	outputEKSAddons, errEKSAddons := eksSvc.ListAddons(context.Background(), inputDescribeAddons)
	require.NoError(t, errEKSAddons)

	// perform the diff
	presenceAddonsMap := make(map[string]bool)
	for _, addon := range outputEKSAddons.Addons {
		presenceAddonsMap[addon] = true
	}
	for _, addonName := range expectedEKSAddons {
		assert.Truef(t, presenceAddonsMap[addonName], "Addon %s not installed on the EKS cluster", addonName)
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
		assert.NoErrorf(t, err, "Failed to get IAM EKS role %s", roleName)
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
	require.NoError(t, errVPC)

	assert.Equal(t, len(outputVPC.Vpcs), 1)

	// key
	keyDescription := fmt.Sprintf("%s -  EKS Secret Encryption Key", clusterName)
	inputKMS := &kms.ListKeysInput{}
	outputKMSList, errKMSList := kmsSvc.ListKeys(context.Background(), inputKMS)
	assert.NoError(t, errKMSList)

	// Check if the key corresponding to the description exists
	keyFound := false
	for _, key := range outputKMSList.Keys {
		keyDetails, errKey := kmsSvc.DescribeKey(context.Background(), &kms.DescribeKeyInput{
			KeyId: key.KeyId,
		})
		require.NoErrorf(t, errKey, "Failed to describe key %s", *key.KeyId)

		keyFound = *keyDetails.KeyMetadata.Description == keyDescription
		if keyFound {
			break
		}
	}
	assert.Truef(t, keyFound, "Failed to find key %s", keyDescription)
}
