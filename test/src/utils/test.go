package utils

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"k8s.io/apimachinery/pkg/util/runtime"
	"testing"
)

func DeferCleanup(t *testing.T, terraformOptions *terraform.Options) {
	fmt.Println("Cleaning up resources")

	sess, err := GetAwsClient()
	if err != nil {
		t.Fatalf("Failed to get AWS client: %v", err)
	}

	// Function to delete objects from S3 bucket
	deleteObjectsFromS3 := func() {
		errDeleteBucket := DeleteObjectFromS3Bucket(sess, terraformOptions.BackendConfig["bucket"].(string), terraformOptions.BackendConfig["key"].(string))
		if errDeleteBucket != nil {
			t.Errorf("Failed to delete objects from S3 bucket: %v", errDeleteBucket)
		}
	}

	destroyTerraform := func() {
		terraform.Destroy(t, terraformOptions)
		deleteObjectsFromS3()
	}

	defer destroyTerraform()
	defer runtime.HandleCrash(func(i interface{}) {
		destroyTerraform()
	})
}
