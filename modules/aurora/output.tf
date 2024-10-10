output "aurora_endpoint" {
  value       = aws_rds_cluster.aurora_cluster.endpoint
  description = "The endpoint of the Aurora cluster"
}

output "aurora_role_name" {
  description = "The name of the aurora IAM role"
  value       = var.iam_create_aurora_role ? aws_iam_role.aurora_role[0].name : ""
  sensitive   = false
}

output "aurora_role_arn" {
  description = "The ARN of the aurora IAM role"
  value       = var.iam_create_aurora_role ? aws_iam_role.aurora_role[0].arn : ""
  sensitive   = false
}

output "aurora_policy_arn" {
  description = "The ARN of the aurora access policy"
  value       = var.iam_create_aurora_role ? aws_iam_policy.aurora_access_policy[0].arn : ""
  sensitive   = false
}
