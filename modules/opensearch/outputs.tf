
output "opensearch_cluster" {
  value       = aws_opensearch_domain.opensearch_cluster
  description = "OpenSearch cluster output"
  sensitive   = true
}

output "opensearch_domain_endpoint" {
  description = "The endpoint of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.endpoint
  sensitive   = false
}

output "opensearch_domain_arn" {
  description = "The ARN of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.arn
  sensitive   = false
}

output "opensearch_domain_id" {
  description = "The ID of the OpenSearch domain"
  value       = aws_opensearch_domain.opensearch_cluster.domain_id
  sensitive   = false
}

output "kms_key_arn" {
  description = "The ARN of the KMS key used to encrypt the OpenSearch domain"
  value       = aws_kms_key.kms.arn
  sensitive   = false
}

output "kms_key_id" {
  description = "The ID of the KMS key used for OpenSearch domain encryption"
  value       = aws_kms_key.kms.key_id
  sensitive   = false
}

output "security_group_id" {
  description = "The ID of the security group used by OpenSearch"
  value       = aws_security_group.this.id
  sensitive   = false
}

output "security_group_rule_ingress" {
  description = "Ingress rule information for OpenSearch security group"
  value       = aws_security_group_rule.allow_ingress
  sensitive   = false
}

output "security_group_rule_egress" {
  description = "Egress rule information for OpenSearch security group"
  value       = aws_security_group_rule.allow_egress
  sensitive   = false
}

output "opensearch_role_name" {
  description = "The name of the OpenSearch IAM role"
  value       = aws_iam_role.opensearch[0].name
  sensitive   = false
}

output "opensearch_role_arn" {
  description = "The ARN of the OpenSearch IAM role"
  value       = aws_iam_role.opensearch[0].arn
  sensitive   = false
}

output "opensearch_policy_arn" {
  description = "The ARN of the OpenSearch access policy"
  value       = aws_iam_policy.opensearch_access_policy[0].arn
  sensitive   = false
}
