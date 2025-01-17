#!/bin/bash

set -euo pipefail

# Count the number of VPCs matching a query name

# Check if a region argument is provided
if [ -z "$1" ]; then
  echo "Error: No region specified."
  echo "Usage: $0 <region>"
  exit 1
fi

REGION=$1
VPC_NAME=$2

# Fetch VPC details based on the name
vpcs=$(aws ec2 describe-vpcs \
  --region "$REGION" \
  --filters "Name=tag:Name,Values=$VPC_NAME" \
  --query 'Vpcs[*].{VpcId:VpcId,CidrBlock:CidrBlock}' \
  --output json)

# Check if the AWS CLI command was successful
if [ $? -ne 0 ]; then
  echo "Error: Failed to fetch VPC data for region $REGION and VPC name $VPC_NAME."
  exit 1
fi

# Parse VPC details and count
vpc_count=$(echo "$vpcs" | jq length)

# Return the VPC data in a format Terraform's external data source expects
echo "{\"vpc_count\": \"$vpc_count\"}"
