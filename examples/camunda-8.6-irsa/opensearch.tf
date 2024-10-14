locals {
  opensearch_domain_name = "domain-name-opensearch" # Replace "domain-name" with your domain name

  # IRSA configuration
  camunda_namespace                = "camunda"     # Replace with your Kubernetes namespace that will host C8 Platform
  camunda_zeebe_service_account    = "zeebe-sa"    # Replace with your Kubernetes ServiceAcccount that will be created for Zeebe
  camunda_operate_service_account  = "operate-sa"  # Replace with your Kubernetes ServiceAcccount that will be created for Operate
  camunda_tasklist_service_account = "tasklist-sa" # Replace with your Kubernetes ServiceAcccount that will be created for TaskList
  camunda_optimize_service_account = "optimize-sa" # Replace with your Kubernetes ServiceAcccount that will be created for Optimize
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

  # Supply your own secret values
  advanced_security_master_user_name     = "secret_user"
  advanced_security_master_user_password = "secretvalue%23"

  depends_on = [module.eks_cluster]

  # IRSA configuration
  iam_create_opensearch_role = true
  iam_opensearch_role_name   = "OpenSearchRole-${local.opensearch_domain_name}" # Ensure uniqueness

  iam_opensearch_access_policy = <<EOF
            {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "es:DescribeElasticsearchDomains",
                    "es:DescribeElasticsearchInstanceTypeLimits",
                    "es:DescribeReservedElasticsearchInstanceOfferings",
                    "es:DescribeReservedElasticsearchInstances",
                    "es:GetCompatibleElasticsearchVersions",
                    "es:ListDomainNames",
                    "es:ListElasticsearchInstanceTypes",
                    "es:ListElasticsearchVersions",
                    "es:DescribeElasticsearchDomain",
                    "es:DescribeElasticsearchDomainConfig",
                    "es:ESHttpGet",
                    "es:ESHttpHead",
                    "es:GetUpgradeHistory",
                    "es:GetUpgradeStatus",
                    "es:ListTags",
                    "es:AddTags",
                    "es:RemoveTags",
                    "es:ESHttpDelete",
                    "es:ESHttpPost",
                    "es:ESHttpPut"
                  ],
                  "Resource": "arn:aws:es:${local.eks_cluster_region}:${module.eks_cluster.aws_caller_identity_account_id}:domain/${local.opensearch_domain_name}/*"
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
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${local.camunda_namespace}:${local.camunda_zeebe_service_account}",
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${local.camunda_namespace}:${local.camunda_operate_service_account}",
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${local.camunda_namespace}:${local.camunda_tasklist_service_account}",
                    "${module.eks_cluster.oidc_provider_id}:sub": "system:serviceaccount:${local.camunda_namespace}:${local.camunda_optimize_service_account}",
                  }
                }
              }
            ]
          }
EOF
}
