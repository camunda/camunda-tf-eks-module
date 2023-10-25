output "cluster_endpoint" {
  description = "Endpoint for your Kubernetes API server"
  value       = module.eks.cluster_endpoint
}

################################################################################
# Security Group
################################################################################

output "cluster_security_group_id" {
  description = "Cluster security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication. Referred to as 'Cluster security group' in the EKS console"
  value       = module.eks.cluster_security_group_id
}

output "cluster_security_group_arn" {
  description = "Amazon Resource Name (ARN) of the cluster security group"
  value       = module.eks.cluster_security_group_arn
}

################################################################################
# IAM Role
################################################################################

output "cluster_iam_role_name" {
  description = "IAM role name of the EKS cluster"
  value       = module.eks.cluster_iam_role_name
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = module.eks.cluster_iam_role_arn
}

# output "eks_admin_iam_role_name" {
#   description = "EKS admin IAM Role name"
#   value       = aws_iam_role.eks_admin_role.name
# }

# output "eks_admin_iam_role_arn" {
#   description = "EKS admin IAM Role arn"
#   value       = aws_iam_role.eks_admin_role.arn
# }

################################################################################
# IRSA
################################################################################

# The following outputs are role arns that are required for the IAM to SA mapping
# Usage: Copy the output value and assign it as an annotation to a SA for usage
# example:
# eks.amazonaws.com/role-arn: arn:aws:iam::831074465991:role/irsa-${RANDOM}
# This allows the SA, if mapping was defined properly, to impersonate the role
output "ebs_cs_arn" {
  value = module.ebs_cs_role.iam_role_arn
}

################################################################################
# VPC
################################################################################

output "vpc_id" {
  description = "VPC id of the cluster"
  value       = module.vpc.vpc_id
}

output "private_vpc_cidr_blocks" {
  value = module.vpc.private_subnets_cidr_blocks
}

output "public_vpc_cidr_blocks" {
  value = module.vpc.public_subnets_cidr_blocks
}

output "private_subnet_ids" {
  value = module.vpc.private_subnets
}
