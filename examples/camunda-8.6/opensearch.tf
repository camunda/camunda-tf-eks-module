locals {
  opensearch_domain_name = "domain-name-os-std" # Replace "domain-name" with your domain name

  opensearch_master_username = "secret_user"    # Replace with your opensearch username
  opensearch_master_password = "Secretvalue$23" # Replace with your opensearch password
}

module "opensearch_domain" {
  source         = "git::https://github.com/camunda/camunda-tf-eks-module//modules/opensearch?ref=2.6.0"
  domain_name    = local.opensearch_domain_name
  engine_version = "2.15"

  instance_type   = "t3.medium.search"
  instance_count  = 3
  ebs_volume_size = 50

  subnet_ids  = module.eks_cluster.private_subnet_ids
  vpc_id      = module.eks_cluster.vpc_id
  cidr_blocks = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)

  advanced_security_enabled                        = true
  advanced_security_internal_user_database_enabled = true

  advanced_security_master_user_name     = local.opensearch_master_username
  advanced_security_master_user_password = local.opensearch_master_password

  depends_on = [module.eks_cluster]
}

output "opensearch_endpoint" {
  value       = module.opensearch_domain.opensearch_domain_endpoint
  description = "The OpenSearch endpoint URL"
}
