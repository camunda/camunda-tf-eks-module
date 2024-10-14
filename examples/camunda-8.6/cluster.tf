locals {
  eks_cluster_name   = "cluster-name-std" # Change this to a name of your choice
  eks_cluster_region = "eu-west-2"        # Change this to your desired AWS region
}

module "eks_cluster" {
  source = "git::https://github.com/camunda/camunda-tf-eks-module//modules/eks-cluster?ref=2.6.0"

  name   = local.eks_cluster_name
  region = local.eks_cluster_region

  # Set CIDR ranges or use the defaults
  cluster_service_ipv4_cidr = "10.190.0.0/16"
  cluster_node_ipv4_cidr    = "10.192.0.0/16"
}
