### Aurora Postgres

module "postgresql" {
  source                     = "./modules/aurora"
  engine_version             = var.postgresql_engine_version
  auto_minor_version_upgrade = false
  cluster_name               = "${var.name}-postgresql"

  username         = var.postgresql_username
  password         = var.postgresql_password
  vpc_id           = module.eks.vpc_id
  subnet_ids       = module.eks.private_subnet_ids
  cidr_blocks      = concat(module.eks.private_vpc_cidr_blocks, module.eks.public_vpc_cidr_blocks)
  instance_class   = "db.t3.medium"
  iam_auth_enabled = true

  depends_on = [module.eks]
}
