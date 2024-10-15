locals {
  aurora_cluster_name = "cluster-name-pg-irsa" # Replace "cluster-name" with your cluster's name

  aurora_master_username = "secret_user"    # Replace with your Aurora username
  aurora_master_password = "secretvalue%23" # Replace with your Aurora password

  camunda_database_keycloak   = "camunda_keycloak"   # Name of your camunda database for Keycloak
  camunda_database_identity   = "camunda_identity"   # Name of your camunda database for Identity
  camunda_database_webmodeler = "camunda_webmodeler" # Name of your camunda database for WebModeler

  # IRSA configuration
  aurora_irsa_keycloak_username   = "keycloak_irsa"   # This is the username that will be used for IRSA connection to the DB on Keycloak db
  aurora_irsa_identity_username   = "identity_irsa"   # This is the username that will be used for IRSA connection to the DB on Identity db
  aurora_irsa_webmodeler_username = "webmodeler_irsa" # This is the username that will be used for IRSA connection to the DB on WebModeler db

  camunda_keycloak_service_account   = "keycloak-sa"   # Replace with your Kubernetes ServiceAcccount that will be created for Keycloak
  camunda_identity_service_account   = "identity-sa"   # Replace with your Kubernetes ServiceAcccount that will be created for Identity
  camunda_webmodeler_service_account = "webmodeler-sa" # Replace with your Kubernetes ServiceAcccount that will be created for WebModeler
}

module "postgresql" {
  source                     = "git::https://github.com/camunda/camunda-tf-eks-module//modules/aurora?ref=2.6.0"
  engine_version             = "15.8"
  auto_minor_version_upgrade = false
  cluster_name               = local.aurora_cluster_name
  default_database_name      = local.camunda_database_keycloak

  availability_zones = ["${local.eks_cluster_region}a", "${local.eks_cluster_region}b", "${local.eks_cluster_region}c"]

  username = local.aurora_master_username
  password = local.aurora_master_password

  vpc_id      = module.eks_cluster.vpc_id
  subnet_ids  = module.eks_cluster.private_subnet_ids
  cidr_blocks = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)

  instance_class = "db.t3.medium"

  # IAM IRSA addition
  iam_aurora_role_name   = "AuroraRole-${local.aurora_cluster_name}" # Ensure this name is unique
  iam_create_aurora_role = true
  iam_auth_enabled       = true

  iam_aurora_access_policy = <<EOF
            {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "rds-db:connect"
                  ],
                  "Resource": [
                    "arn:aws:rds-db:${local.eks_cluster_region}:${module.eks_cluster.aws_caller_identity_account_id}:dbuser:${local.aurora_cluster_name}/${local.aurora_irsa_keycloak_username}",
                    "arn:aws:rds-db:${local.eks_cluster_region}:${module.eks_cluster.aws_caller_identity_account_id}:dbuser:${local.aurora_cluster_name}/${local.aurora_irsa_identity_username}",
                    "arn:aws:rds-db:${local.eks_cluster_region}:${module.eks_cluster.aws_caller_identity_account_id}:dbuser:${local.aurora_cluster_name}/${local.aurora_irsa_webmodeler_username}"
                  ]
                }
              ]
            }
EOF

  iam_role_trust_policy = <<EOF
          {
            "Version": "2012-10-17",
            "Statement": [
              {
                "Effect": "Allow",
                "Principal": {
                  "Federated": "${module.eks_cluster.oidc_provider_arn}"
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
                "Condition": {
                  "StringEquals": {
                    "${module.eks_cluster.oidc_provider_id}:sub": [
                      "system:serviceaccount:${local.camunda_namespace}:${local.camunda_webmodeler_service_account}",
                      "system:serviceaccount:${local.camunda_namespace}:${local.camunda_identity_service_account}",
                      "system:serviceaccount:${local.camunda_namespace}:${local.camunda_keycloak_service_account}"
                    ]
                  }
                }
              }
            ]
          }
EOF

  depends_on = [module.eks_cluster]
}

output "postgres_endpoint" {
  value       = module.postgresql.aurora_endpoint
  description = "The Postgres endpoint URL"
}

output "aurora_role_arn" {
  value       = module.postgresql.aurora_role_arn
  description = "The Aurora Role ARN used for IRSA"
}
