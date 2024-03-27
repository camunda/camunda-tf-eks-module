package test

// adapted from https://github.com/cloudposse/terraform-aws-eks-cluster/blob/main/test/src/examples_complete_test.go

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newClientSet(cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        aws.StringValue(cluster.Endpoint),
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

// Function to check a specific VPC attribute
func checkVpcAttribute(ec2Svc *ec2.EC2, vpcID, attributeName string) (*ec2.DescribeVpcAttributeOutput, error) {
	input := &ec2.DescribeVpcAttributeInput{
		VpcId:     aws.String(vpcID),
		Attribute: aws.String(attributeName),
	}

	return ec2Svc.DescribeVpcAttribute(input)
}

// Test the Terraform module in modules/eks-cluster.
func TestModulesEKSCluster(t *testing.T) {

	randId := strings.ToLower(random.UniqueId())
	clusterName := fmt.Sprintf("cluster-%s", randId)
	region := "eu-central-1"

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/eks-cluster",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.eu-central-1.eks.tfvars"},
		Vars: map[string]interface{}{
			"name":   clusterName,
			"region": region,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
	defer runtime.HandleCrash(func(i interface{}) {
		terraform.Destroy(t, terraformOptions)
	})

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	// then it will re-run apply to make sure that out tf is idempotent
	_, errTfApply := terraform.InitAndApplyAndIdempotentE(t, terraformOptions)
	assert.NoError(t, errTfApply)

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
	expectedPrivateVpcCidrBlocks := []string([]string{"10.192.0.0/19", "10.192.32.0/19", "10.192.64.0/19"})
	assert.Equal(t, terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"), expectedPrivateVpcCidrBlocks)

	// this is a split(6)[3..5] of the base cluster_node_ipv4_cidr    = "10.192.0.0/16"
	expectedPublicVpcCidrBlocks := []string([]string{"10.192.96.0/19", "10.192.128.0/19", "10.192.160.0/19"})
	assert.Equal(t, terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"), expectedPublicVpcCidrBlocks)

	// Wait for the worker nodes to join the cluster
	// https://github.com/kubernetes/client-go
	// https://www.rushtehrani.com/post/using-kubernetes-api
	// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap
	// https://gianarb.it/blog/kubernetes-shared-informer
	// https://stackoverflow.com/questions/60547409/unable-to-obtain-kubeconfig-of-an-aws-eks-cluster-in-go-code/60573982#60573982
	fmt.Println("Waiting for worker nodes to join the EKS cluster")

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	// list your services here
	eksSvc := eks.New(sess)
	iamSvc := iam.New(sess)
	ec2Svc := ec2.New(sess)
	kmsSvc := kms.New(sess)

	inputEKS := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}

	result, err := eksSvc.DescribeCluster(inputEKS)
	assert.NoError(t, err)

	clientSet, err := newClientSet(result.Cluster)
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
			if countOfWorkerNodes > 1 {
				close(stopChannel)
			}
		},
	})
	require.NoError(t, errEventHandler)

	go informer.Run(stopChannel)

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
	inputDescribeAddons := &eks.ListAddonsInput{}
	outputEKSAddons, errEKSAddons := eksSvc.ListAddons(inputDescribeAddons)
	require.NoError(t, errEKSAddons)

	// perform the diff
	presenceAddonsMap := make(map[string]bool)
	for _, addon := range outputEKSAddons.Addons {
		presenceAddonsMap[*addon] = true
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

		_, err := iamSvc.GetRole(input)
		assert.NoErrorf(t, err, "Failed to get IAM EKS role %s", roleName)
	}

	// verifies the VPC

	vpcName := fmt.Sprintf("%s-vpc", clusterName)

	inputVPC := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(vpcName)},
			},
		},
	}

	outputVPC, errVPC := ec2Svc.DescribeVpcs(inputVPC)
	require.NoError(t, errVPC)

	assert.Equal(t, len(outputVPC.Vpcs), 1)

	vpc := outputVPC.Vpcs[0]
	vpcID := *vpc.VpcId
	assert.NotEmpty(t, vpcID)

	// WIP
	/*	val, err := checkVpcAttribute(ec2Svc, vpcID, "mapPublicIpOnLaunch")
		assert.NoError(t, err)
		assert.Equal(t, true, *val)
		checkVpcAttribute(ec2Svc, vpcID, "enableDnsSupport")
		checkVpcAttribute(ec2Svc, vpcID, "enableDnsHostnames")
	*/

	keyDescription := fmt.Sprintf("%s -  EKS Secret Encryption Key", clusterName)
	inputKMS := &kms.ListKeysInput{}
	outputKMSList, errKMSList := kmsSvc.ListKeys(inputKMS)
	assert.NoError(t, errKMSList)

	// Check if the key corresponding to the description exists
	keyFound := false
	for _, key := range outputKMSList.Keys {
		keyDetails, err := kmsSvc.DescribeKey(&kms.DescribeKeyInput{
			KeyId: key.KeyId,
		})
		require.NoErrorf(t, err, "Failed to describe key %s", *key.KeyId)

		keyFound = *keyDetails.KeyMetadata.Description == keyDescription
		if keyFound {
			break
		}
	}
	assert.Truef(t, keyFound, "Failed to find key %s", keyDescription)
}
