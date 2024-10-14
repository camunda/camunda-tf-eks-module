locals {
  opensearch_domain_name = "domain-name-opensearch" # Replace "domain-name" with your domain name
}

module "opensearch" {
  source         = "git::https://github.com/camunda/camunda-tf-eks-module//modules/opensearch?ref=2.6.0"
  domain_name    = local.opensearch_domain_name
  engine_version = "2.15"

  instance_type   = "t3.medium.search"
  instance_count  = 3
  ebs_volume_size = 50

  subnet_ids         = module.eks_cluster.private_subnet_ids
  security_group_ids = module.eks_cluster.security_group_ids
  vpc_id             = module.eks_cluster.vpc_id
  cidr_blocks        = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)

  advanced_security_enabled                        = true
  advanced_security_internal_user_database_enabled = true

  # Supply your own secret values
  advanced_security_master_user_name     = "secret_user"
  advanced_security_master_user_password = "secretvalue%23"

  depends_on = [module.eks_cluster]
}
