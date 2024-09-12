# Deploy RDS Aurora Cluster GitHub Action

This GitHub Action automates the deployment of an Amazon RDS Aurora cluster using Terraform. It installs Terraform and AWS CLI, and outputs the Aurora cluster endpoint along with other relevant details.

## Description

The **Deploy RDS Aurora Cluster** action enables you to:

- Automate the deployment of an RDS Aurora cluster on AWS.
- Use Terraform for infrastructure as code.
- Install specific versions of Terraform and AWS CLI.
- Output the Aurora cluster endpoint, Terraform state URL, and all other Terraform outputs dynamically.

## Inputs

The following inputs are required for the action:

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `aws-region` | AWS region where the RDS Aurora cluster will be deployed. | Yes | - |
| `cluster-name` | Name of the RDS Aurora cluster to deploy. | Yes | - |
| `engine-version` | Version of the Aurora engine to use. | Yes | see `action.yml` |
| `instance-class` | Instance class for the Aurora cluster. | Yes | `db.t3.medium` |
| `num-instances` | Number of instances in the Aurora cluster. | Yes | `1` |
| `username` | Username for the PostgreSQL admin user. | Yes | - |
| `password` | Password for the PostgreSQL admin user. | Yes | - |
| `vpc-id` | VPC ID to create the cluster in. | No | - |
| `subnet-ids` | List of subnet IDs to create the cluster in. | No | `[]` |
| `cidr-blocks` | CIDR blocks to allow access from and to. | No | `[]` |
| `s3-backend-bucket` | Name of the S3 bucket to store Terraform state. | Yes | - |
| `s3-bucket-region` | Region of the bucket containing the resources states. Fallbacks to `aws-region` if not set. | No | - |
| `tf-modules-revision` | Git revision of the Terraform modules to use. | Yes | `main` |
| `tf-modules-path` | Path where the Terraform Aurora modules will be cloned. | Yes | `./.action-tf-modules/aurora/` |
| `tf-cli-config-credentials-hostname` | The hostname of a HCP Terraform/Terraform Enterprise instance for the CLI configuration file. | No | `app.terraform.io` |
| `tf-cli-config-credentials-token` | The API token for a HCP Terraform/Terraform Enterprise instance. | No | - |
| `tf-terraform-version` | The version of Terraform CLI to install. | No | `latest` |
| `tf-terraform-wrapper` | Whether to install a wrapper for the Terraform binary. | No | `true` |
| `awscli-version` | Version of the AWS CLI to use. | Yes | see `action.yml` |

## Outputs

The action provides the following outputs:

| Output | Description |
|--------|-------------|
| `aurora-endpoint` | The endpoint of the deployed Aurora cluster. |
| `terraform-state-url` | URL of the Terraform state file in the S3 bucket. |
| `all-terraform-outputs` | All outputs from Terraform. |

## Usage

To use this GitHub Action, include it in your workflow file:

```yaml
jobs:
  deploy_aurora:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy Aurora Cluster
        uses: camunda/camunda-tf-eks-module/aurora-manage-cluster@main
        with:
          aws-region: 'us-west-2'
          cluster-name: 'my-aurora-cluster'
          engine-version: '15.4'
          instance-class: 'db.t3.medium'
          num-instances: '2'
          username: 'admin'
          password: ${{ secrets.DB_PASSWORD }}
          vpc-id: 'vpc-12345678'
          subnet-ids: 'subnet-12345,subnet-67890'
          cidr-blocks: '10.0.0.0/16'
          tags: '{"env": "prod", "team": "devops"}'
          s3-backend-bucket: 'my-terraform-state-bucket'
          s3-bucket-region: 'us-west-2'
```
