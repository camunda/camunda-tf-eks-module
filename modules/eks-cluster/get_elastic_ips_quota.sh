#!/bin/bash

set -euo pipefail

# Check if a region argument is provided
if [ -z "$1" ]; then
  echo "Error: No region specified."
  echo "Usage: $0 <region>"
  exit 1
fi

REGION=$1

# Fetch the Elastic IP quota for the specified region using AWS CLI
quota=$(aws service-quotas get-service-quota --region "$REGION" --service-code ec2 --quota-code L-0263D0A3 --query 'Quota.Value' --output text)

# Check if the AWS CLI command for quota was successful
if [ $? -ne 0 ]; then
  echo "Error: Failed to fetch Elastic IP quota for region $REGION."
  exit 1
fi

# Return the quota value in a format Terraform's external data source expects (string: string)
echo "{\"quota\": \"$quota\"}"
