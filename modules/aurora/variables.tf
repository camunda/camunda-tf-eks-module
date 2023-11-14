variable "cluster_name" {
  description = "Name of the cluster, also used to prefix dependent resources. Format: /[[:lower:][:digit:]-]/"
}

variable "engine" {
  default     = "aurora-postgresql"
  description = "The engine type e.g. aurora, aurora-mysql, aurora-postgresql, ..."
}

variable "engine_version" {
  type        = string
  default     = "15.4"
  description = "The DB engine version for Postgres to use."
}

variable "auto_minor_version_upgrade" {
  default     = true
  description = "If true, minor engine upgrades will be applied automatically to the DB instance during the maintenance window"
}

variable "availability_zones" {
  type        = list(string)
  default     = ["eu-central-1a", "eu-central-1b", "eu-central-1c"]
  description = "Array of zones"
}

variable "instance_class" {
  default = "db.r5.large"
}

variable "num_instances" {
  default     = "1"
  description = "Number of instances"
}

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

variable "tags" {
  default = {}
}

variable "subnet_ids" {
  type = list(string)
}

variable "cidr_blocks" {
  type = list(string)
}

# pass through from root
variable "vpc_id" {
}

# Allows adding additional iam roles to grant access from Aurora to e.g. S3
variable "iam_roles" {
  default = []
}

# Allows defining whether IAM auth should be activated for IRSA usage
variable "iam_auth_enabled" {
  default = false
  type    = bool
}

variable "ca_cert_identifier" {
  default = "rds-ca-2019"
  type    = string
}
