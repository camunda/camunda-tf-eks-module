locals {
  vpc_name = "${var.name}-vpc"
}

locals {
  # Generate the list of availability zones
  azs = var.availability_zones != null ? var.availability_zones : [
    for index in range(var.availability_zones_count) : "${var.region}${["a", "b", "c", "d", "e", "f"][index]}"
  ]

  # Private subnets for nodes
  private_subnets = [
    for index in range(length(local.azs)) : cidrsubnet(var.cluster_node_ipv4_cidr, length(local.azs), index)
  ]

  public_subnets = [
    for index in range(length(local.azs)) : cidrsubnet(var.cluster_node_ipv4_cidr, length(local.azs), index + length(local.azs))
  ]
}


module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.16.0"

  name = local.vpc_name
  # requires a /16 range, resulting in 2 leftover subnets, which can be used for DBs
  # AWS supports between /16 and 28
  cidr = var.cluster_node_ipv4_cidr

  azs = local.azs

  # Private subnets for nodes
  private_subnets = local.private_subnets

  # Public subnets for Load Balancers
  public_subnets = local.public_subnets

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }
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
