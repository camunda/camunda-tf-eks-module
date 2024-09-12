# Deploy or Destroy EKS Cluster

This GitHub Action automates the deployment or destruction of an Amazon Elastic Kubernetes Service (EKS) cluster using Terraform. It also installs necessary tools like Terraform, AWS CLI, and `kubectl`, and sets up the Kubernetes context for the created cluster.

## Usage

To use this action, add it to your workflow file (e.g., `.github/workflows/eks-deploy.yml`):

```yaml
name: EKS Cluster Management

on:
  workflow_dispatch:

jobs:
  eks_management:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy or Destroy EKS Cluster
        uses: camunda/camunda-tf-eks-module/eks-manage-cluster@main
        with:
          action: 'create' # or 'destroy'
          aws-region: 'us-west-2'
          cluster-name: 'my-eks-cluster'
          kubernetes-version: '1.30'
          cluster-service-ipv4-cidr: '10.190.0.0/16'
          cluster-node-ipv4-cidr: '10.192.0.0/16'
          np-instance-types: '["t2.medium"]'
          np-capacity-type: 'SPOT'
          np-node-desired-count: '4'
          np-node-min-count: '1'
          np-disk-size: '20'
          np-ami-type: 'AL2_x86_64'
          np-node-max-count: '10'
          s3-backend-bucket: 'your-terraform-state-bucket'
          s3-bucket-region: 'us-west-2'
          tf-modules-revision: 'main'
          tf-modules-path: './.action-tf-modules/eks/'
          login: 'true'
          awscli-version: '2.15.52'
```

## Inputs

| Input Name                          | Description                                                                                                  | Required | Default                          |
|-------------------------------------|--------------------------------------------------------------------------------------------------------------|----------|----------------------------------|
| `aws-region`                        | AWS region where the EKS cluster will be deployed.                                                           | Yes      | N/A                              |
| `cluster-name`                      | Name of the EKS cluster to deploy.                                                                           | Yes      | N/A                              |
| `kubernetes-version`                | Version of Kubernetes to use for the EKS cluster.                                                            | Yes      | `1.30`                           |
| `cluster-service-ipv4-cidr`         | CIDR block for cluster service IPs.                                                                          | Yes      | `10.190.0.0/16`                  |
| `cluster-node-ipv4-cidr`            | CIDR block for cluster node IPs.                                                                             | Yes      | `10.192.0.0/16`                  |
| `np-instance-types`                 | List of instance types for the node pool.                                                                    | Yes      | `["t2.medium"]`                  |
| `np-capacity-type`                  | Capacity type for non-production instances (e.g., SPOT).                                                     | Yes      | `SPOT`                           |
| `np-node-desired-count`             | Desired number of nodes in the EKS node group.                                                               | Yes      | `4`                              |
| `np-node-min-count`                 | Minimum number of nodes in the EKS node group.                                                               | Yes      | `1`                              |
| `np-disk-size`                      | Disk size of the nodes on the default node pool (in GB).                                                     | Yes      | `20`                             |
| `np-ami-type`                       | Amazon Machine Image type.                                                                                   | Yes      | `AL2_x86_64`                     |
| `np-node-max-count`                 | Maximum number of nodes in the EKS node group.                                                               | Yes      | `10`                             |
| `s3-backend-bucket`                 | Name of the S3 bucket to store Terraform state.                                                              | Yes      | N/A                              |
| `s3-bucket-region`                  | Region of the bucket containing the resources states; falls back on `aws-region` if not set.                 | No       | N/A                              |
| `tf-modules-revision`               | Git revision of the Terraform modules to use.                                                                | Yes      | `main`                           |
| `tf-modules-path`                   | Path where the Terraform EKS modules will be cloned.                                                         | Yes      | `./.action-tf-modules/eks/`      |
| `login`                             | Authenticate the current kube context on the created cluster.                                                | Yes      | `true`                           |
| `tf-cli-config-credentials-hostname`| The hostname of a HCP Terraform/Terraform Enterprise instance to use for credentials configuration.           | No       | `app.terraform.io`               |
| `tf-cli-config-credentials-token`   | The API token for a HCP Terraform/Terraform Enterprise instance.                                             | No       | N/A                              |
| `tf-terraform-version`              | The version of Terraform CLI to install. Accepts full version or constraints like `<1.13.0` or `latest`.     | No       | `latest`                         |
| `tf-terraform-wrapper`              | Whether or not to install a wrapper for Terraform CLI calls.                                                 | No       | `true`                           |
| `awscli-version`                    | Version of the AWS CLI to install.                                                                           | Yes      | see `action.yml`                        |

## Outputs

| Output Name                | Description                                                      |
|----------------------------|------------------------------------------------------------------|
| `eks-cluster-endpoint`     | The API endpoint of the deployed EKS cluster.                    |
| `terraform-state-url`      | URL of the Terraform state file in the S3 bucket.                |
| `all-terraform-outputs`    | All outputs from Terraform.                                      |
