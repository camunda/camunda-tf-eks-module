# Deploy an EKS Cluster

## Description

This GitHub Action automates the deployment of an EKS (Amazon Elastic Kubernetes Service) cluster using Terraform.
This action will also install Terraform, awscli, and kubectl. The kube context will be set on the created cluster.


## Inputs

| name | description | required | default |
| --- | --- | --- | --- |
| `aws-region` | <p>AWS region where the EKS cluster will be deployed</p> | `true` | `""` |
| `cluster-name` | <p>Name of the EKS cluster to deploy</p> | `true` | `""` |
| `kubernetes-version` | <p>Version of Kubernetes to use for the EKS cluster</p> | `false` | `1.30` |
| `cluster-service-ipv4-cidr` | <p>CIDR block for cluster service IPs</p> | `false` | `10.190.0.0/16` |
| `cluster-node-ipv4-cidr` | <p>CIDR block for cluster node IPs</p> | `false` | `10.192.0.0/16` |
| `np-instance-types` | <p>List of instance types</p> | `false` | `["t2.medium"]` |
| `np-capacity-type` | <p>Capacity type for non-production instances (e.g., SPOT)</p> | `false` | `SPOT` |
| `np-node-desired-count` | <p>Desired number of nodes in the EKS node group</p> | `false` | `4` |
| `np-node-min-count` | <p>Minimum number of nodes in the EKS node group</p> | `false` | `1` |
| `np-disk-size` | <p>Disk size of the nodes on the default node pool</p> | `false` | `20` |
| `np-ami-type` | <p>Amazon Machine Image</p> | `false` | `AL2_x86_64` |
| `np-node-max-count` | <p>Maximum number of nodes in the EKS node group</p> | `false` | `10` |
| `s3-backend-bucket` | <p>Name of the S3 bucket to store Terraform state</p> | `true` | `""` |
| `s3-bucket-region` | <p>Region of the bucket containing the resources states, if not set, will fallback on aws-region</p> | `false` | `""` |
| `tf-modules-revision` | <p>Git revision of the tf modules to use</p> | `false` | `main` |
| `tf-modules-path` | <p>Path where the tf EKS modules will be cloned</p> | `false` | `./.action-tf-modules/eks/` |
| `login` | <p>Authenticate the current kube context on the created cluster</p> | `false` | `true` |
| `tf-cli-config-credentials-hostname` | <p>The hostname of a HCP Terraform/Terraform Enterprise instance to place within the credentials block of the Terraform CLI configuration file. Defaults to <code>app.terraform.io</code>.</p> | `false` | `app.terraform.io` |
| `tf-cli-config-credentials-token` | <p>The API token for a HCP Terraform/Terraform Enterprise instance to place within the credentials block of the Terraform CLI configuration file.</p> | `false` | `""` |
| `tf-terraform-version` | <p>The version of Terraform CLI to install. Instead of full version string you can also specify constraint string starting with "&lt;" (for example <code>&lt;1.13.0</code>) to install the latest version satisfying the constraint. A value of <code>latest</code> will install the latest version of Terraform CLI. Defaults to <code>latest</code>.</p> | `false` | `latest` |
| `tf-terraform-wrapper` | <p>Whether or not to install a wrapper to wrap subsequent calls of the <code>terraform</code> binary and expose its STDOUT, STDERR, and exit code as outputs named <code>stdout</code>, <code>stderr</code>, and <code>exitcode</code> respectively. Defaults to <code>true</code>.</p> | `false` | `true` |
| `awscli-version` | <p>Version of the aws cli to use</p> | `false` | `2.15.52` |


## Outputs

| name | description |
| --- | --- |
| `eks-cluster-endpoint` | <p>The API endpoint of the deployed EKS cluster</p> |
| `terraform-state-url` | <p>URL of the Terraform state file in the S3 bucket</p> |
| `all-terraform-outputs` | <p>All outputs from Terraform</p> |


## Runs

This action is a `composite` action.

## Usage

```yaml
- uses: ***PROJECT***@***VERSION***
  with:
    aws-region:
    # AWS region where the EKS cluster will be deployed
    #
    # Required: true
    # Default: ""

    cluster-name:
    # Name of the EKS cluster to deploy
    #
    # Required: true
    # Default: ""

    kubernetes-version:
    # Version of Kubernetes to use for the EKS cluster
    #
    # Required: false
    # Default: 1.30

    cluster-service-ipv4-cidr:
    # CIDR block for cluster service IPs
    #
    # Required: false
    # Default: 10.190.0.0/16

    cluster-node-ipv4-cidr:
    # CIDR block for cluster node IPs
    #
    # Required: false
    # Default: 10.192.0.0/16

    np-instance-types:
    # List of instance types
    #
    # Required: false
    # Default: ["t2.medium"]

    np-capacity-type:
    # Capacity type for non-production instances (e.g., SPOT)
    #
    # Required: false
    # Default: SPOT

    np-node-desired-count:
    # Desired number of nodes in the EKS node group
    #
    # Required: false
    # Default: 4

    np-node-min-count:
    # Minimum number of nodes in the EKS node group
    #
    # Required: false
    # Default: 1

    np-disk-size:
    # Disk size of the nodes on the default node pool
    #
    # Required: false
    # Default: 20

    np-ami-type:
    # Amazon Machine Image
    #
    # Required: false
    # Default: AL2_x86_64

    np-node-max-count:
    # Maximum number of nodes in the EKS node group
    #
    # Required: false
    # Default: 10

    s3-backend-bucket:
    # Name of the S3 bucket to store Terraform state
    #
    # Required: true
    # Default: ""

    s3-bucket-region:
    # Region of the bucket containing the resources states, if not set, will fallback on aws-region
    #
    # Required: false
    # Default: ""

    tf-modules-revision:
    # Git revision of the tf modules to use
    #
    # Required: false
    # Default: main

    tf-modules-path:
    # Path where the tf EKS modules will be cloned
    #
    # Required: false
    # Default: ./.action-tf-modules/eks/

    login:
    # Authenticate the current kube context on the created cluster
    #
    # Required: false
    # Default: true

    tf-cli-config-credentials-hostname:
    # The hostname of a HCP Terraform/Terraform Enterprise instance to place within the credentials block of the Terraform CLI configuration file. Defaults to `app.terraform.io`.
    #
    # Required: false
    # Default: app.terraform.io

    tf-cli-config-credentials-token:
    # The API token for a HCP Terraform/Terraform Enterprise instance to place within the credentials block of the Terraform CLI configuration file.
    #
    # Required: false
    # Default: ""

    tf-terraform-version:
    # The version of Terraform CLI to install. Instead of full version string you can also specify constraint string starting with "<" (for example `<1.13.0`) to install the latest version satisfying the constraint. A value of `latest` will install the latest version of Terraform CLI. Defaults to `latest`.
    #
    # Required: false
    # Default: latest

    tf-terraform-wrapper:
    # Whether or not to install a wrapper to wrap subsequent calls of the `terraform` binary and expose its STDOUT, STDERR, and exit code as outputs named `stdout`, `stderr`, and `exitcode` respectively. Defaults to `true`.
    #
    # Required: false
    # Default: true

    awscli-version:
    # Version of the aws cli to use
    #
    # Required: false
    # Default: 2.15.52
```
