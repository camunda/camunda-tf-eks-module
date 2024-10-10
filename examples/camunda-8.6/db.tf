locals {
  aurora_cluster_name = "cluster-name-postgresql" # Replace "cluster-name" with your cluster's name
}

module "postgresql" {
  source                     = "git::https://github.com/camunda/camunda-tf-eks-module//modules/aurora?ref=feature/opensearch"
  engine_version             = "15.8"
  auto_minor_version_upgrade = false
  cluster_name               = locals.aurora_cluster_name
  default_database_name      = "camunda"

  # Supply your own secret values for username and password
  username = "secret_user"
  password = "secretvalue%23"

  vpc_id      = module.eks_cluster.vpc_id
  subnet_ids  = module.eks_cluster.private_subnet_ids
  cidr_blocks = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)

  instance_class   = "db.t3.medium"
  iam_auth_enabled = true

  depends_on = [module.eks_cluster]
}
