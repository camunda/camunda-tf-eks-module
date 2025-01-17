#!/bin/bash

set -euo pipefail

# Check if a region argument is provided
if [ -z "$1" ]; then
  echo "Error: No region specified."
  echo "Usage: $0 <region>"
  exit 1
fi

REGION=$1

# Fetch all Elastic IPs for the specified region using AWS CLI
eips=$(aws ec2 describe-addresses --region "$REGION" --query 'Addresses[*].{PublicIp:PublicIp,AllocationId:AllocationId,InstanceId:InstanceId}' --output json)

# Check if the AWS CLI command was successful
if [ $? -ne 0 ]; then
  echo "Error: Failed to fetch Elastic IPs for region $REGION."
  exit 1
fi

eips_count=$(echo "$eips" | jq length)

# Return the quota value in a format Terraform's external data source expects (string: string)
echo "{\"elastic_ips_count\": \"$eips_count\"}"
