
output "opensearch_cluster" {
  value       = aws_opensearch_domain.opensearch_cluster
  description = "OpenSearch cluster output"
}

output "opensearch_domain_endpoint" {
  description = "The endpoint of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.endpoint
}

output "opensearch_domain_arn" {
  description = "The ARN of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.arn
}

output "opensearch_domain_id" {
  description = "The ID of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.domain_id
}

output "kms_key_arn" {
  description = "The ARN of the KMS key used to encrypt the OpenSearch domain"
  value       = aws_kms_key.key.arn
}

output "kms_key_id" {
  description = "The ID of the KMS key used for OpenSearch domain encryption"
  value       = aws_kms_key.key.key_id
}

output "security_group_id" {
  description = "The ID of the security group used by OpenSearch"
  value       = aws_security_group.this.id
}

output "security_group_rule_ingress" {
  description = "Ingress rule information for OpenSearch security group"
  value       = aws_security_group_rule.allow_ingress
}

output "security_group_rule_egress" {
  description = "Egress rule information for OpenSearch security group"
  value       = aws_security_group_rule.allow_egress
}
