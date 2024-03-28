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
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sync/atomic"
	"testing"
	"time"
)

func TestTearsUp(t *testing.T) {
	// log
	logger := zaptest.NewLogger(t)
	sugar := logger.Sugar()

	//randId := strings.ToLower(random.UniqueId())
	randId := "leo"
	clusterName := fmt.Sprintf("cluster-%s", randId)
	region := "eu-central-1"
	sugar.Infow("Creating EKS cluster...")

	terraformOptions := TearsUpEKS(t, sugar, clusterName, region, "")

	// test suite
	testEksCluster(t, sugar, terraformOptions)

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	/*	defer terraform.Destroy(t, terraformOptions)

		// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
		defer runtime.HandleCrash(func(i interface{}) {
			terraform.Destroy(t, terraformOptions)
		})*/

	TearsDown(t, sugar)
}

// TearsUpEKS spawns a new EKS Cluster with a random name from a fixture file
func TearsUpEKS(t *testing.T, sugar *zap.SugaredLogger, clusterName, region, kubernetesVersion string) *terraform.Options {
	varsConfig := map[string]interface{}{
		"name":   clusterName,
		"region": region,
	}

	if kubernetesVersion != "" {
		varsConfig["kubernetes_version"] = kubernetesVersion
	}

	sugar.Infow("TF vars", varsConfig)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/eks-cluster",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.eu-central-1.eks.tfvars"},
		Vars:     varsConfig,
	}

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	// then it will re-run apply to make sure that out tf is idempotent
	terraform.InitAndApplyAndIdempotent(t, terraformOptions)
	return terraformOptions
}

func TearsDown(t *testing.T, sugar *zap.SugaredLogger) {
	sugar.Infow("Tests completed")
}

// Test the Terraform module in modules/eks-cluster.
func testEksCluster(t *testing.T, sugar *zap.SugaredLogger, terraformOptions *terraform.Options) {
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

	assert.Equal(t, fmt.Sprintf("%s-eks-iam-role", clusterName), terraform.Output(t, terraformOptions, "cluster_iam_role_name"))

	// this is a split(6)[0..2] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPrivateVpcCidrBlocks := "[10.192.0.0/19 10.192.32.0/19 10.192.64.0/19]"
	assert.Equal(t, expectedPrivateVpcCidrBlocks, terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"))

	// this is a split(6)[3..5] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPublicVpcCidrBlocks := "[10.192.96.0/19 10.192.128.0/19 10.192.160.0/19]"
	assert.Equal(t, expectedPublicVpcCidrBlocks, terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"))

	// Wait for the worker nodes to join the cluster
	// https://github.com/kubernetes/client-go
	// https://www.rushtehrani.com/post/using-kubernetes-api
	// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap
	// https://gianarb.it/blog/kubernetes-shared-informer
	// https://stackoverflow.com/questions/60547409/unable-to-obtain-kubeconfig-of-an-aws-eks-cluster-in-go-code/60573982#60573982
	fmt.Println("Waiting for worker nodes to join the EKS cluster")

	sess, err := utils.GetAwsClient()
	require.NoErrorf(t, err, "Failed to get aws client")

	// list your services here
	eksSvc := eks.NewFromConfig(sess)
	iamSvc := iam.NewFromConfig(sess)
	ec2Svc := ec2.NewFromConfig(sess)
	kmsSvc := kms.NewFromConfig(sess)

	var expectedNodesCount uint64 = 3

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(context.TODO(), inputEKS)
	assert.NoError(t, err)

	clientSet, err := utils.NewClientSet(result.Cluster)
	assert.NoError(t, err)

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	informer := factory.Core().V1().Nodes().Informer()
	stopChannel := make(chan struct{})
	var countOfWorkerNodes uint64 = 0

	_, errEventHandler := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*corev1.Node)
			fmt.Printf("Worker Node %s has joined the EKS cluster at %s\n", node.Name, node.CreationTimestamp)
			atomic.AddUint64(&countOfWorkerNodes, 1)
			if countOfWorkerNodes >= expectedNodesCount {
				stopChannel <- struct{}{} // send close signal
			}
		},
	})
	require.NoError(t, errEventHandler)

	go informer.Run(stopChannel)
	go func() {
		// wait to receive a signal to close the channel
		<-stopChannel
		close(stopChannel)
	}()

	select {
	case <-stopChannel:
		msg := "All worker nodes have joined the EKS cluster"
		fmt.Println(msg)
	case <-time.After(5 * time.Minute):
		msg := "Not all worker nodes have joined the EKS cluster"
		fmt.Println(msg)
		assert.Fail(t, msg)
	}

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

	vpc := outputVPC.Vpcs[0]
	vpcID := *vpc.VpcId
	assert.NotEmpty(t, vpcID)

	// todo: implement
	/*	valAttr, errAttr := utils.CheckVpcAttribute(ec2Svc, vpcID, types.VpcAttributeNameEnableDnsHostnames)
		assert.NoError(t, errAttr)
		assert.Equal(t, true, *valAttr)
		valAttr, errAttr = utils.CheckVpcAttribute(ec2Svc, vpcID, types.VpcAttributeNameEnableDnsSupport)
		assert.NoError(t, errAttr)
		assert.Equal(t, true, *valAttr)*/

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
