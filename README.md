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

```hcl

module "opensearch_domain" {
  source = "github.com/camunda/camunda-tf-eks-module/modules/opensearch"

  domain_name     = "my-opensearch-domain"
  subnet_ids      = module.eks_cluster.private_subnet_ids
  security_group_ids = module.eks_cluster.security_group_ids
  vpc_id          = module.eks_cluster.vpc_id
  cidr_blocks      = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)

  instance_type   = "t3.small.search"
  instance_count  = 4
  ebs_volume_size = 100

  advanced_security_enabled = true
  advanced_security_internal_user_database_enabled = true
  advanced_security_master_user_name = "admin"
  advanced_security_master_user_password = "password"

  depends_on = [module.eks_cluster]
}
```

#### GitHub Actions

You can automate the deployment and deletion of the EKS cluster and Aurora database using GitHub Actions. Below are examples of GitHub Actions workflows for deploying and deleting these resources.

For more details, refer to the corresponding [EKS Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/eks-manage-cluster/README.md), [Aurora Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/aurora-manage-cluster/README.md) and [OpenSearch Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/opensearch-manage-cluster/README.md), [Cleanup Actions README](https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/actions/eks-cleanup-resources/README.md).

An example workflow can be found in https://github.com/camunda/camunda-tf-eks-module/blob/main/.github/workflows/test-gha-eks.yml.

#### Advanced usage with IRSA

This documentation provides a step-by-step guide to creating an EKS cluster, an Aurora RDS instance, and an OpenSearch domain with IRSA (IAM Roles for Service Accounts) roles using Terraform modules.
The modules create the necessary IAM roles and policies for Aurora and OpenSearch. To simplify the configuration, the modules use the outputs of the EKS cluster module to define the IRSA roles and policies.

### Aurora IRSA Role and Policy

The Aurora module uses the following outputs from the EKS cluster module to define the IRSA role and policy:

- `module.eks_cluster.oidc_provider_arn`: The ARN of the OIDC provider for the EKS cluster.
- `module.eks_cluster.oidc_provider_id`: The ID of the OIDC provider for the EKS cluster.
- `var.account_id`: Your account id
- `var.aurora_cluster_name`: The name of the Aurora cluster to access
- `var.aurora_irsa_username`: The username of the user used to access to the AuroraDB
- `var.aurora_namespace`: The namespace to allow access
- `var.aurora_service_account`: The ServiceAccount to allow access

You need to define the IAM role trust policy and access policy for Aurora. Here's an example of how to define these policies using the outputs of the EKS cluster module:

```hcl
module "postgresql" {
  # ...
  iam_aurora_access_policy = <<EOF
            {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "rds-db:connect"
                  ],
                  "Resource": "arn:aws:rds-db:${module.eks_cluster.region}:${var.account_id}:dbuser:${var.aurora_cluster_name}/${var.aurora_irsa_username}"
                }
              ]
            }
EOF

  iam_role_trust_policy = <<EOF
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Allow",
                "Principal": {
                  "Federated": "${module.eks_cluster.oidc_provider_arn}"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                  "StringEquals": {
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${var.aurora_namespace}:${var.aurora_service_account}"
                  }
                }
              }
            ]
          }
EOF

  iam_aurora_role_name = "AuroraRole-your-cluster" # ensure uniqueness of this one
  iam_create_aurora_role = true
  iam_auth_enabled = true
  # ...
}
```

### OpenSearch IRSA Role and Policy

The OpenSearch module uses the following outputs from the EKS cluster module to define the IRSA role and policy:

- `module.eks_cluster.oidc_provider_arn`: The ARN of the OIDC provider for the EKS cluster.
- `module.eks_cluster.oidc_provider_id`: The ID of the OIDC provider for the EKS cluster.
- `var.account_id`: Your account id
- `var.opensearch_domain_name`: The name of the OpenSearch domain to access
- `var.opensearch_namespace`: The namespace to allow access
- `var.opensearch_service_account`: The ServiceAccount to allow access

You need to define the IAM role trust policy and access policy for OpenSearch. Here's an example of how to define these policies using the outputs of the EKS cluster module:

```hcl
module "opensearch_domain" {
  # ...
  iam_create_opensearch_role = true
  iam_opensearch_role_name = "OpenSearchRole-your-cluster" # ensure uniqueness of this one
  iam_opensearch_access_policy = <<EOF
            {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "es:ESHttpGet",
                    "es:ESHttpPut",
                    "es:ESHttpPost"
                  ],
                  "Resource": "arn:aws:es:${module.eks_cluster.region}:${var.account_id}:domain/${var.opensearch_domain_name}/*"
                }
              ]
            }
EOF

  iam_role_trust_policy = <<EOF
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Allow",
                "Principal": {
                  "Federated": "${module.eks_cluster.oidc_provider_arn}"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                  "StringEquals": {
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${var.opensearch_namespace}:${var.opensearch_service_account}"
                  }
                }
              }
            ]
          }
EOF
  # ...
}
```

By defining the IRSA roles and policies using the outputs of the EKS cluster module, you can simplify the configuration and ensure that the roles and policies are created with the correct permissions and trust policies.

## Support

Please note that the modules have been tested with **[Terraform](https://github.com/hashicorp/terraform)** in the version described in the [.tool-versions](./.tool-versions) of this project.
