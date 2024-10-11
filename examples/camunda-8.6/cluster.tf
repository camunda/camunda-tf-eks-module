module "eks_cluster" {
  source = "git::https://github.com/camunda/camunda-tf-eks-module//modules/eks-cluster?ref=2.6.0"

  region = "eu-west-2"    # Change this to your desired AWS region
  name   = "cluster-name" # Change this to a name of your choice

  # Set CIDR ranges or use the defaults
  cluster_service_ipv4_cidr = "10.190.0.0/16"
  cluster_node_ipv4_cidr    = "10.192.0.0/16"
}
