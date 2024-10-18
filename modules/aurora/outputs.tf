output "aurora_endpoint" {
  value       = aws_rds_cluster.aurora_cluster.endpoint
  description = "The endpoint of the Aurora cluster"
}

// Output for Role ARNs
output "aurora_iam_role_arns" {
  description = "Map of IAM role names to their ARNs"

  value = { for role_name, role in aws_iam_role.roles : role_name => role.arn }
  sensitive   = false
}

// Output for Policy ARNs
output "aurora_iam_role_access_policy_arns" {
  description = "Map of IAM role names to their access policy ARNs"

  value = { for role_name, policy in aws_iam_policy.access_policies : role_name => policy.arn }

  sensitive   = false
}