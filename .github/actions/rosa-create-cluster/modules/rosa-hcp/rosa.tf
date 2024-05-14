locals {
  account_role_prefix  = "${var.cluster_name}-account"
  operator_role_prefix = "${var.cluster_name}-operator"

  tags = {
    "owner" = data.aws_caller_identity.current.arn
  }
}


module "rosa_hcp" {
  source  = "terraform-redhat/rosa-hcp/rhcs"
  version = "1.6.2-prerelease.1"

  openshift_version = var.openshift_version
  cluster_name      = var.cluster_name

  compute_machine_type = var.compute_node_instance_type
  tags                 = local.tags

  machine_cidr = module.vpc.cidr_block
  properties   = { rosa_creator_arn = data.aws_caller_identity.current.arn }


  replicas               = var.replicas
  aws_availability_zones = module.vpc.availability_zones
  aws_subnet_ids         = concat(module.vpc.public_subnets, module.vpc.private_subnets)

  host_prefix = var.host_prefix

  // STS configuration
  create_account_roles  = true
  account_role_prefix   = local.account_role_prefix
  create_oidc           = true
  create_operator_roles = true
  operator_role_prefix  = local.operator_role_prefix

  wait_for_create_complete            = true
  wait_for_std_compute_nodes_complete = true

  depends_on = [module.vpc]
}

module "htpasswd_idp" {
  source  = "terraform-redhat/rosa-hcp/rhcs//modules/idp"
  version = "1.6.2-prerelease.1"

  cluster_id         = module.rosa_hcp.cluster_id
  name               = "htpasswd-idp"
  idp_type           = "htpasswd"
  htpasswd_idp_users = [{ username = var.htpasswd_username, password = var.htpasswd_password }]
}

module "vpc" {
  source  = "terraform-redhat/rosa-hcp/rhcs//modules/vpc"
  version = "1.6.2-prerelease.1"

  name_prefix              = var.cluster_name
  availability_zones_count = var.availability_zones_count

  vpc_cidr = var.vpc_cidr_block
}
