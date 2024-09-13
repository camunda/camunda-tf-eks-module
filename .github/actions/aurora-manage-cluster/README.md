# Deploy RDS Aurora Cluster GitHub Action

This GitHub Action automates the deployment of an Amazon RDS Aurora cluster using Terraform. It installs Terraform and AWS CLI, and outputs the Aurora cluster endpoint along with other relevant details.

## Description

The **Deploy RDS Aurora Cluster** action enables you to:

- Automate the deployment of an RDS Aurora cluster on AWS.
- Use Terraform for infrastructure as code.
- Install specific versions of Terraform and AWS CLI.
- Output the Aurora cluster endpoint, Terraform state URL, and all other Terraform outputs dynamically.

## Inputs

The following inputs are required or optional for the action:

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `cluster-name` | Name of the RDS Aurora cluster to deploy. | Yes | - |
| `engine-version` | Version of the Aurora engine to use. | Yes | see `action.yml` |
| `instance-class` | Instance class for the Aurora cluster. | Yes | `db.t3.medium` |
| `num-instances` | Number of instances in the Aurora cluster. | Yes | `1` |
| `username` | Username for the PostgreSQL admin user. | Yes | - |
| `password` | Password for the PostgreSQL admin user. | Yes | - |
| `vpc-id` | VPC ID to create the cluster in. | Yes | - |
| `subnet-ids` | List of subnet IDs to create the cluster in. | Yes | - |
| `cidr-blocks` | CIDR blocks to allow access from and to. | Yes | - |
| `auto-minor-version-upgrade` | If true, minor engine upgrades will be applied automatically to the DB instance during the maintenance window. | No | `true` |
| `availability-zones` | Array of availability zones to use for the Aurora cluster. | No | `[]` |
| `iam-roles` | Allows propagating additional IAM roles to the Aurora cluster for features like access to S3. | No | `[]` |
| `iam-auth-enabled` | Determines whether IAM authentication should be activated for IRSA usage. | No | `false` |
| `ca-cert-identifier` | Specifies the identifier of the CA certificate for the DB instance. | No | `rds-ca-rsa2048-g1` |
| `default-database-name` | The name for the automatically created database on cluster creation. | No | `camunda` |
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
          cluster-name: 'my-aurora-cluster'
          engine-version: '15.4'
          instance-class: 'db.t3.medium'
          num-instances: '2'
          username: 'admin'
          password: ${{ secrets.DB_PASSWORD }}
          vpc-id: 'vpc-12345678'
          subnet-ids: 'subnet-12345,subnet-67890'
          cidr-blocks: '10.0.0.0/16'
          auto-minor-version-upgrade: 'true'
          availability-zones: '["us-west-2a", "us-west-2b"]'
          iam-roles: '["arn:aws:iam::123456789012:role/my-role"]'
          iam-auth-enabled: 'false'
          ca-cert-identifier: 'rds-ca-rsa2048-g1'
          default-database-name: 'mydatabase'
          s3-backend-bucket: 'my-terraform-state-bucket'
          s3-bucket-region: 'us-west-2'
          tf-modules-revision: 'main'
          tf-modules-path: './.action-tf-modules/aurora/'
          awscli-version: '2.15.52'
```
