



variable "username" {
  type        = string
  description = "The username for the postgres admin user. Important: secret value!"
  sensitive   = true
}

variable "password" {
  type        = string
  description = "The password for the postgres admin user. Important: secret value!"
  sensitive   = true
}

variable "subnet_ids" {
  type        = list(string)
  description = "The subnet IDs to create the cluster in. For easier usage we are passing through the subnet IDs from the AWS EKS Cluster module."
}

# pass through from root
variable "vpc_id" {
  description = "The VPC ID to create the cluster in. For easier usage we are passing through the VPC ID from the AWS EKS Cluster module."
}

# Allows defining whether IAM auth should be activated for IRSA usage
variable "iam_auth_enabled" {
  default     = false
  type        = bool
  description = "Determines whether IAM auth should be activated for IRSA usage"
}

variable "default_database_name" {
  type        = string
  default     = "camunda"
  description = "The name for the automatically created database on cluster creation."
}
