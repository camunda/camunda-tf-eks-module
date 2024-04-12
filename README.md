# Camunda Terraform EKS Modules

Terraform module which creates AWS EKS (Kubernetes) resources with an opinionated configuration targeting Camunda 8.

## Documentation

The related [guide](https://docs.camunda.io/docs/next/self-managed/setup/deploy/amazon/amazon-eks/eks-terraform/) describing a more detailed usage.
Consider installing Camunda 8 via [following guide](https://docs.camunda.io/docs/next/self-managed/setup/deploy/amazon/amazon-eks/eks-helm/) after having deployed the AWS EKS cluster.

## Usage

Following is a simple example configuration and should be adjusted as required.

See [AWS EKS Cluster inputs](./modules/eks-cluster/README.md#inputs) and [AWS Aurora RDS inputs](./modules/aurora/README.md#inputs) for further configuration options and how they affect the cluster creation.

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

## Support

Please note that the modules have been tested with **[OpenTofu](https://opentofu.org/)** in the version described in the [.tool-versions](./.tool-versions) of this project.
