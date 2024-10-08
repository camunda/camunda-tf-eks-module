# Camunda Terraform EKS Modules

[![Camunda](https://img.shields.io/badge/Camunda-FC5D0D)](https://www.camunda.com/)
[![tests](https://github.com/camunda/camunda-tf-eks-module/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/camunda/camunda-tf-eks-module/actions/workflows/tests.yml)
[![License](https://img.shields.io/github/license/camunda/camunda-tf-eks-module)](LICENSE)

Terraform module which creates AWS EKS (Kubernetes) resources with an opinionated configuration targeting Camunda 8 and an AWS Aurora RDS cluster.

**⚠️ Warning:** This project is not intended for production use but rather for demonstration purposes only. There are no guarantees or warranties provided.

## Documentation

The related [guide](https://docs.camunda.io/docs/next/self-managed/setup/deploy/amazon/amazon-eks/eks-terraform/) describes more detailed usage.
Consider installing Camunda 8 via [this guide](https://docs.camunda.io/docs/next/self-managed/setup/deploy/amazon/amazon-eks/eks-helm/) after deploying the AWS EKS cluster.

## Usage

Below is a simple example configuration for deploying both an EKS cluster and an Aurora PostgreSQL database.

See [AWS EKS Cluster inputs](./modules/eks-cluster/README.md#inputs) and [AWS Aurora RDS inputs](./modules/aurora/README.md#inputs) for further configuration options and how they affect the cluster and database creation.

```hcl
module "eks_cluster" {
  source = "github.com/camunda/camunda-tf-eks-module/modules/eks-cluster"

  region             = "eu-central-1"
  name               = "cluster-name"

  cluster_service_ipv4_cidr = "10.190.0.0/16"
  cluster_node_ipv4_cidr    = "10.192.0.0/16"
}
```

```hcl
module "postgresql" {
  source                     = "github.com/camunda/camunda-tf-eks-module/modules/aurora"
  engine_version             = "15.4"
  auto_minor_version_upgrade = false
  cluster_name               = "cluster-name-postgresql"

  username         = "username"
  password         = "password"
  vpc_id           = module.eks_cluster.vpc_id
  subnet_ids       = module.eks_cluster.private_subnet_ids
  cidr_blocks      = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)
  instance_class   = "db.t3.medium"
  iam_auth_enabled = true

  depends_on = [module.eks_cluster]
}
```

#### GitHub Actions

You can automate the deployment and deletion of the EKS cluster and Aurora database using GitHub Actions. Below are examples of GitHub Actions workflows for deploying and deleting these resources.

For more details, refer to the corresponding [EKS Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/eks-manage-cluster/README.md) and [Aurora Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/aurora-manage-cluster/README.md), [Cleanup Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/eks-cleanup-resources/README.md).

An example workflow can be found in https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/workflows/test-gha-eks.yml.

## Support

Please note that the modules have been tested with **[Terraform](https://github.com/hashicorp/terraform)** in the version described in the [.tool-versions](./.tool-versions) of this project.
