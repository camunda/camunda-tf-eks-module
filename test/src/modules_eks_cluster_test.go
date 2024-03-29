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
	"github.com/camunda/camunda-tf-eks-module/utils"
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
	"os/exec"
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
	region := "eu-central-1"
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
	region := "eu-central-1"
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

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.TODO(), inputEKS)
	assert.NoError(t, err)

	sugar.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilClusterIsReady(result.Cluster, 5*time.Minute, uint64(expectedCapacity))
	require.NoError(t, errClusterReady)

	// Spawn RDS within the EKS VPC/subnet
	publicBlocks := strings.Fields(strings.Trim(terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"), "[]"))
	privateBlocks := strings.Fields(strings.Trim(terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"), "[]"))

	auroraUsername := "myuser"
	auroraPassword := "mypassword123secure"

	varsConfigAurora := map[string]interface{}{
		"username":         auroraUsername,
		"password":         auroraPassword,
		"cluster_name":     fmt.Sprintf("postgres-%s", clusterSuffix),
		"subnet_ids":       result.Cluster.ResourcesVpcConfig.SubnetIds,
		"vpc_id":           *result.Cluster.ResourcesVpcConfig.VpcId,
		"cidr_blocks":      append(publicBlocks, privateBlocks...),
		"iam_auth_enabled": true,
	}

	terraformOptionsRDS := SpawnAurora(t, sugar, varsConfigAurora)
	auroraEndpoint := terraform.Output(t, terraformOptionsRDS, "aurora_endpoint")
	assert.NotEmpty(t, auroraEndpoint)

	// Launch a pod on the cluster and test the RDS
	kubeClient, err := utils.NewClientSet(result.Cluster)
	require.NoError(t, err)

	// create kubeconfig
	cmd := exec.Command("aws", "eks", "--region", region, "update-kubeconfig", "--name", clusterName, "--profile", utils.GetAwsProfile(), "--kubeconfig", "kubeconfig")
	_, errCmdKubeProfile := cmd.Output()
	require.NoError(t, errCmdKubeProfile)

	// create the configmap
	namespace := "postgres-client"
	pgKubeCtlOptions := k8s.NewKubectlOptions("", "kubeconfig", namespace)
	_, errFindNamespace := k8s.GetNamespaceE(t, pgKubeCtlOptions, namespace)
	if errFindNamespace != nil {
		if errors.IsNotFound(err) {
			k8s.CreateNamespace(t, pgKubeCtlOptions, namespace)
		} else {
			require.NoError(t, errFindNamespace)
		}
	}

	// todo: https: //github.com/camunda/c8-multi-region/blob/main/test/internal/helpers/aws/helpers.go#L242

	configMapPostgres := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aurora-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"aurora_endpoint": auroraEndpoint,
			"aurora_username": auroraUsername,
			"aurora_port":     "5432",
			"aws_region":      region,
			"aurora_db_name":  auroraEndpoint,
		},
	}
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), configMapPostgres, metav1.CreateOptions{})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			require.NoError(t, err)
		}
	}
	// cleanup existing jobs
	jobListOptions := metav1.ListOptions{LabelSelector: "app=postgres-client"}
	existingJobs := k8s.ListJobs(t, pgKubeCtlOptions, jobListOptions)
	for _, job := range existingJobs {
		err := kubeClient.BatchV1().Jobs(namespace).Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
		assert.NoError(t, err)
	}

	k8s.KubectlApply(t, pgKubeCtlOptions, "../../test/src/fixtures/postgres-client.yml")

	errJob := utils.WaitForJobCompletion(kubeClient, namespace, "postgres-client", 5*time.Minute, jobListOptions)
	require.NoError(t, errJob)

	TearsDown(t, sugar)
}

// TestUpgradeEKS starts from a version of EKS, deploy a simple chart, upgrade the cluster
// and check that everything is working as expected
func TestUpgradeEKS(t *testing.T) {
	// log
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	clusterSuffix := utils.GetEnv("TESTS_CLUSTER_ID", strings.ToLower(random.UniqueId()))
	clusterName := fmt.Sprintf("cluster-upgrade-%s", clusterSuffix)
	region := "eu-central-1"
	sugar.Infow("Creating EKS cluster...")
	expectedCapacity := 4

	varsConfig := map[string]interface{}{
		"name":                  clusterName,
		"region":                region,
		"np_desired_node_count": expectedCapacity,
		"kubernetes_version":    "1.27",
	}

	terraformOptions := SpawnEKS(t, sugar, varsConfig)

	// test suite
	baseChecksEKS(t, sugar, terraformOptions, 3)

	// upgrade the cluster
	varsConfig["kubernetes_version"] = "1.28"

	terraformOptions = SpawnEKS(t, sugar, varsConfig)

	// check everything works as expected
	baseChecksEKS(t, sugar, terraformOptions, 3)

	TearsDown(t, sugar)
}

// SpawnEKS spawns a new EKS Cluster from a default fixture file
func SpawnEKS(t *testing.T, sugar *zap.SugaredLogger, varsConfig map[string]interface{}) *terraform.Options {
	sugar.Infow("TF vars", varsConfig)

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
	sugar.Infow("TF vars", varsConfig)

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
	sugar.Infow("Testing status of the EKS cluster", clusterName)

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

	result, err := eksSvc.DescribeCluster(context.TODO(), inputEKS)
	assert.NoError(t, err)

	// Wait for the worker nodes to join the cluster
	sugar.Infow("Waiting for worker nodes to join the EKS cluster")
	errClusterReady := utils.WaitUntilClusterIsReady(result.Cluster, 5*time.Minute, expectedNodesCount)
	require.NoError(t, errClusterReady)

	// Verify list of addons installed on the EKS
	expectedEKSAddons := []string{"coredns", "kube-proxy", "vpc-cni", "aws-ebs-csi-driver"}
	inputDescribeAddons := &eks.ListAddonsInput{
		ClusterName: aws.String(clusterName),
	}
	outputEKSAddons, errEKSAddons := eksSvc.ListAddons(context.TODO(), inputDescribeAddons)
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

		_, err := iamSvc.GetRole(context.TODO(), input)
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

	outputVPC, errVPC := ec2Svc.DescribeVpcs(context.TODO(), inputVPC)
	require.NoError(t, errVPC)

	assert.Equal(t, len(outputVPC.Vpcs), 1)

	// key
	keyDescription := fmt.Sprintf("%s -  EKS Secret Encryption Key", clusterName)
	inputKMS := &kms.ListKeysInput{}
	outputKMSList, errKMSList := kmsSvc.ListKeys(context.TODO(), inputKMS)
	assert.NoError(t, errKMSList)

	// Check if the key corresponding to the description exists
	keyFound := false
	for _, key := range outputKMSList.Keys {
		keyDetails, errKey := kmsSvc.DescribeKey(context.TODO(), &kms.DescribeKeyInput{
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

// todo: test upgrade path
// todo: test auroradb integration
// todo: test inputs
