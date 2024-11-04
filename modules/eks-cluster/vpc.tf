locals {
  vpc_name = "${var.name}-vpc"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.15.0"

  name = local.vpc_name
  # requires a /16 range, resulting in 2 leftover subnets, which can be used for DBs
  # AWS supports between /16 and 28
  cidr = var.cluster_node_ipv4_cidr

  azs = ["${var.region}a", "${var.region}b", "${var.region}c"]

  # private subnets for nodes
  private_subnets = [cidrsubnet(var.cluster_node_ipv4_cidr, 3, 0), cidrsubnet(var.cluster_node_ipv4_cidr, 3, 1), cidrsubnet(var.cluster_node_ipv4_cidr, 3, 2)]
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }

  # public subnet for loadbalancers
  public_subnets = [cidrsubnet(var.cluster_node_ipv4_cidr, 3, 3), cidrsubnet(var.cluster_node_ipv4_cidr, 3, 4), cidrsubnet(var.cluster_node_ipv4_cidr, 3, 5)]
  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }

  # Don't assign public IPv4 addresses on EC2 instance launch
  map_public_ip_on_launch = false

  # Single NATGateway per private subnet
  enable_nat_gateway = true
  single_nat_gateway = false
  reuse_nat_ips      = false

  # Enable DNS hostnames and DNS support required for VPC Peering
  # enable_dns_hostnames = true
  # enable_dns_support   = true

  # Logs IP traffic for whole VPC
  enable_flow_log                      = false
  create_flow_log_cloudwatch_iam_role  = false
  create_flow_log_cloudwatch_log_group = false

}
