package test

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/runtime"
	"testing"
)

// Test the Terraform module in modules/eks-cluster using Terratest.
func TestModulesEKSCluster(t *testing.T) {

	/*	randId := strings.ToLower(random.UniqueId())
		attributes := []string{randId}*/

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../modules/eks-cluster",
		Upgrade:      false,
		// Variables to pass to our Terraform code using -var-file options
		VarFiles: []string{"../../test/src/fixtures/fixtures.eu-central-1.eks.tfvars"},
		Vars:     map[string]interface{}{},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// If Go runtime crushes, run `terraform destroy` to clean up any resources that were created
	defer runtime.HandleCrash(func(i interface{}) {
		terraform.Destroy(t, terraformOptions)
	})

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	errTfApply, _ := terraform.InitAndApplyE(t, terraformOptions)
	require.Nil(t, errTfApply)
}
