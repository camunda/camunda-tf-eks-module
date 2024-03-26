package test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/runtime"
	"strings"
	"testing"
)

// Test the Terraform module in modules/eks-cluster using Terratest.
func TestModulesEKSCluster(t *testing.T) {

	randId := strings.ToLower(random.UniqueId())
	clusterName := fmt.Sprintf("cluster-%s", randId)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/eks-cluster",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.eu-central-1.eks.tfvars"},
		Vars: map[string]interface{}{
			"name": clusterName,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
	defer runtime.HandleCrash(func(i interface{}) {
		terraform.Destroy(t, terraformOptions)
	})

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	_, errTfApply := terraform.InitAndApplyE(t, terraformOptions)
	require.Nil(t, errTfApply)

	// Do some basic not empty tests on outputs
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_endpoint"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_security_group_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_primary_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_iam_role_name"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "cluster_iam_role_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "ebs_cs_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "external_dns_arn"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "vpc_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_vpc_cidr_blocks"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "public_vpc_cidr_blocks"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_subnet_ids"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "default_security_group_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "vpc_main_route_table_id"))
	assert.NotEmpty(t, terraform.Output(t, terraformOptions, "private_route_table_ids"))
}
