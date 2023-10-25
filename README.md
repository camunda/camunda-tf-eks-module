# Camunda TF EKS Module

Terraform module which creates AWS EKS (Kubernetes) resources with an opiniated configuration targeting Camunda 8.

> [!WARNING]  
> Do not use for production purposes.

## Documentation

TODO: Link to external docs - docs.camunda.io

## Usage

Following is a simple example configuration and should be adjusted as required.

See [inputs](#inputs) for further configuration options and how they affect the cluster creation.

```hcl
module "eks_cluster" {
  source = "github.com/camunda/camunda-tf-eks-module"

  region             = "eu-central-1"
  name               = "cluster-name"

  cluster_service_ipv4_cidr = "10.190.0.0/16"
  cluster_node_ipv4_cidr    = "10.192.0.0/16"
}
```

### Inputs

TODO: generate something with tf docs
