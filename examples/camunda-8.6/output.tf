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
