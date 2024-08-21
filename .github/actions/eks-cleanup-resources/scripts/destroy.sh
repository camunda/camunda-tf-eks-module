#!/bin/bash

# Description:
# This script performs a Terraform destroy operation for resources defined in an S3 bucket.
# It copies the Terraform module directory to a temporary location, initializes Terraform with
# the appropriate backend configuration, and runs `terraform destroy`. If the destroy operation
# is successful, it removes the corresponding S3 objects.
#
# Usage:
# ./destroy_resources.sh <BUCKET> <MODULES_DIR> <TEMP_DIR_PREFIX> <MIN_AGE_IN_HOURS> <ID_OR_ALL>
#
# Arguments:
#   BUCKET: The name of the S3 bucket containing the resource state files.
#   MODULES_DIR: The directory containing the Terraform modules.
#   TEMP_DIR_PREFIX: The prefix for the temporary directories created for each resource.
#   MIN_AGE_IN_HOURS: The minimum age (in hours) of resources to be destroyed.
#   ID_OR_ALL: The specific ID suffix to filter objects, or "all" to destroy all objects.
#
# Example:
# ./destroy.sh tf-state-eks-ci-eu-west-3 ./modules/eks/ /tmp/eks/ 24 all
# ./destroy.sh tf-state-eks-ci-eu-west-3 ./modules/eks/ /tmp/eks/ 24 4891048
#
# Requirements:
# - AWS CLI installed and configured with the necessary permissions to access and modify the S3 bucket.
# - Terraform installed and accessible in the PATH.

# Check for required arguments
if [ "$#" -ne 5 ]; then
  echo "Usage: $0 <BUCKET> <MODULES_DIR> <TEMP_DIR_PREFIX> <MIN_AGE_IN_HOURS> <ID_OR_ALL>"
  exit 1
fi

if [ -z "$AWS_REGION" ]; then
  echo "Error: The environment variable AWS_REGION is not set."
  exit 1
fi

# Detect operating system and set the appropriate date command
if [[ "$(uname)" == "Darwin" ]]; then
    date_command="gdate"
else
    date_command="date"
fi

# Variables
BUCKET=$1
MODULES_DIR=$2
TEMP_DIR_PREFIX=$3
MIN_AGE_IN_HOURS=$4
ID_OR_ALL=$5
FAILED=0
CURRENT_DIR=$(pwd)

# Function to perform terraform destroy
destroy_resource() {
  local resource_id=$1
  local terraform_module=$2
  local resource_id_dir
  resource_id_dir=$(dirname "$resource_id")
  local temp_dir="${TEMP_DIR_PREFIX}${resource_id_dir}"
  local resource_module_path="$MODULES_DIR$terraform_module/"

  echo "Copying $resource_module_path in $temp_dir"

  mkdir -p "$temp_dir" || return 1
  cp -a "$resource_module_path." "$temp_dir" || return 1

  tree "$resource_module_path" "$temp_dir" || return 1

  cd "$temp_dir" || return 1

  tree "." || return 1

  if ! terraform init -backend-config="bucket=$BUCKET" -backend-config="key=${resource_id}" -backend-config="region=$AWS_REGION"; then return 1; fi

  # Execute the terraform destroy command with appropriate variables (see https://github.com/hashicorp/terraform/issues/23552)
  if [ "$terraform_module" == "eks-cluster" ]; then
    # disable the refresh as it causes errors with kubernetes provider and is not needed to destroy things
    if ! terraform destroy -refresh=false -auto-approve \
      -var="region=$AWS_REGION" \
      -var="name=dummy" \
      -var="cluster_service_ipv4_cidr=10.190.0.0/16" \
      -var="cluster_node_ipv4_cidr=10.192.0.0/16"; then return 1; fi

  elif [ "$terraform_module" == "aurora" ]; then
    if ! terraform destroy -auto-approve \
      -var="cluster_name=dummy" \
      -var="username=dummy" \
      -var="password=dummy" \
      -var="subnet_ids=[]" \
      -var="cidr_blocks=[]" \
      -var="vpc_id=vpc-dummy"; then return 1; fi
  else
    echo "Unsupported module: $terraform_module"
    return 1
  fi

  # Cleanup S3
  echo "Deleting s3://$BUCKET/$resource_id"
  if ! aws s3 rm "s3://$BUCKET/$resource_id" --recursive; then return 1; fi
  if ! aws s3api delete-object --bucket "$BUCKET" --key "$resource_id"; then return 1; fi

  cd - || return 1
  rm -rf "$temp_dir" || return 1
}

# List objects in the S3 bucket and parse the resource IDs
if [ "$ID_OR_ALL" == "all" ]; then
  resources=$(aws s3 ls "s3://$BUCKET/" --recursive | grep "/terraform.tfstate" | awk '{print $4}')
else
  resources=$(aws s3 ls "s3://$BUCKET/" --recursive | grep "/terraform.tfstate" | grep "$ID_OR_ALL" | awk '{print $4}')
fi

current_timestamp=$($date_command +%s)

for resource_id in $resources; do
  cd "$CURRENT_DIR" || return 1

  terraform_module=$(basename "$(dirname "$resource_id")")
  echo "Checking resource $resource_id (terraform module=$terraform_module)"

  last_modified=$(aws s3api head-object --bucket "$BUCKET" --key "$resource_id" --output json | grep LastModified | awk -F '"' '{print $4}')
  if [ -z "$last_modified" ]; then
    echo "Error: Failed to retrieve last modified timestamp for resource $resource_id"
    exit 1
  fi

  last_modified_timestamp=$($date_command -d "$last_modified" +%s)
  if [ -z "$last_modified_timestamp" ]; then
    echo "Error: Failed to convert last modified timestamp to seconds since epoch for resource $resource_id"
    exit 1
  fi
  echo "resource $resource_id last modification: $last_modified ($last_modified_timestamp)"

  file_age_hours=$(( ($current_timestamp - $last_modified_timestamp) / 3600 ))
  if [ -z "$file_age_hours" ]; then
    echo "Error: Failed to calculate file age in hours for resource $resource_id"
    exit 1
  fi
  echo "resource $resource_id is $file_age_hours hours old"

  if [ $file_age_hours -ge "$MIN_AGE_IN_HOURS" ]; then
    echo "Destroying resource $resource_id in $terraform_module"

    if ! destroy_resource "$resource_id" "$terraform_module"; then
      echo "Error destroying resource $resource_id"
      FAILED=1
    fi
  else
    echo "Skipping resource $resource_id as it does not meet the minimum age requirement of $MIN_AGE_IN_HOURS hours"
  fi
done

# Exit with the appropriate status
if [ $FAILED -ne 0 ]; then
  echo "One or more operations failed."
  exit 1
else
  echo "All operations completed successfully."
  exit 0
fi
