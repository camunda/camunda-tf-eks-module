### Aurora Postgres

module "postgresql" {
  source                     = "./modules/aurora"
  engine_version             = var.postgresql_engine_version
  auto_minor_version_upgrade = false
  cluster_name               = "${var.name}-postgresql"

  username         = var.postgresql_username
  password         = var.postgresql_password
  vpc_id           = module.vpc.vpc_id
  subnet_ids       = module.vpc.private_subnets
  cidr_blocks      = concat(module.vpc.private_subnets_cidr_blocks, module.vpc.public_subnets_cidr_blocks)
  instance_class   = "db.t3.medium"
  iam_auth_enabled = true

  depends_on = [module.vpc]
}
