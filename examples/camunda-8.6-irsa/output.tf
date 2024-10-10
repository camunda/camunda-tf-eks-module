output "cert_manager_arn" {
  value       = module.eks_cluster.cert_manager_arn
  description = "The Amazon Resource Name (ARN) of the AWS IAM Roles for Service Account mapping for the cert-manager"
}

output "external_dns_arn" {
  value       = module.eks_cluster.external_dns_arn
  description = "The Amazon Resource Name (ARN) of the AWS IAM Roles for Service Account mapping for the external-dns"
}

output "postgres_endpoint" {
  value       = module.postgresql.aurora_endpoint
  description = "The Postgres endpoint URL"
}

output "opensearch_endpoint" {
  value       = module.opensearch.opensearch_domain_endpoint
  description = "The OpenSearch endpoint URL"
}

output "aurora_role_arn" {
  value       = module.postgresql.aurora_role_arn
  description = "The Aurora Role ARN used for IRSA"
}

output "opensearch_role_arn" {
  value       = module.opensearch.opensearch_role_arn
  description = "The OpenSearch Role ARN used for IRSA"
}
