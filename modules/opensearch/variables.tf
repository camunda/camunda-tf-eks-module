# ! Developer: if you are adding a variable without a default value, please ensure to reference it in the cleanup script (.github/actions/eks-cleanup-resources/scripts/destroy.sh)

variable "domain_name" {
  type        = string
  description = "Name of the domain."
  required    = true
}

variable "engine_version" {
  type        = string
  description = "OpenSearch version for the domain."
  # TODO: add renovate
  default = "2.15"
}

variable "subnet_ids" {
  type        = list(string)
  description = "The subnet IDs to create the cluster in. For easier usage we are passing through the subnet IDs from the AWS EKS Cluster module."
  required    = true
}

variable "cidr_blocks" {
  type        = list(string)
  description = "The CIDR blocks to allow access from and to."
}

variable "vpc_id" {
  type        = string
  description = "VPC used by the domain."
  required    = true
}

variable "availability_zones" {
  type        = list(string)
  description = "Availability zones used by the domain (should match the VPC)."
  required    = true
}

variable "security_group_ids" {
  type        = list(string)
  description = "Additional security groups used by the domain."
  default     = []
}

variable "instance_type" {
  type        = string
  default     = "t3.small.search"
  description = "Instance type of data nodes in the cluster."
}

variable "instance_count" {
  type        = number
  default     = 1
  description = "Number of instances in the cluster."
}

variable "cold_storage_enabled" {
  type        = bool
  default     = false
  description = "Indicates cold storage is enabled."
}

variable "dedicated_master_enabled" {
  type        = bool
  description = "Indicates whether dedicated master nodes are enabled for the cluster."
  default     = true
}

variable "dedicated_master_type" {
  type        = string
  description = "Instance type of the dedicated master nodes in the cluster."
  default     = ""
}

variable "dedicated_master_count" {
  type        = number
  description = "Number of dedicated master nodes in the cluster."
  default     = 1
}

variable "multi_az_with_standby_enabled" {
  type        = bool
  description = "Whether a multi-AZ domain is turned on with a standby AZ."
  default     = false
}

variable "zone_awareness_enabled" {
  type        = bool
  description = "Indicates whether zone awareness is enabled."
  default     = true
}

variable "zone_awareness_availability_zone_count" {
  type        = number
  description = "Number of availability zones used."
  default     = 1
}

variable "warm_enabled" {
  type        = bool
  description = "Warm storage is enabled."
  default     = true
}

variable "warm_count" {
  type        = number
  description = "Number of warm nodes in the cluster."
  default     = 1
}

variable "warm_type" {
  type        = string
  description = "Instance type for the OpenSearch cluster's warm nodes."
  default     = ""
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Tags assigned to the domain."
}

variable "auto_software_update_enabled" {
  type        = bool
  default     = false
  description = "Software update auto for the domain."
}

variable "automated_snapshot_start_hour" {
  type        = number
  default     = 0
  description = "Hour during which the service takes an automated daily snapshot of the indices in the domain."
}

variable "node_to_node_encryption_enabled" {
  type        = bool
  default     = true
  description = "Whether node to node encryption is enabled."
}

variable "advanced_options" {
  type = map(any)
  default = {
    "rest.action.multi.allow_explicit_index" = true
  }
  description = "Key-value string pairs to specify advanced configuration options."
}

variable "advanced_security_enabled" {
  type        = bool
  default     = false
  description = "Whether advanced security is enabled."
}

variable "advanced_security_internal_user_database_enabled" {
  type        = bool
  default     = false
  description = "Whether the internal user database is enabled."
}

variable "advanced_security_master_user_name" {
  type        = string
  default     = "opensearch-admin"
  description = "Main user's username, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `advanced_security_internal_user_database_enabled` is set to true."
}

variable "advanced_security_master_user_password" {
  type        = string
  description = "Main user's password, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `advanced_security_internal_user_database_enabled` is set to true."
}

variable "advanced_security_anonymous_auth_enabled" {
  type        = bool
  description = "Whether the anonymous auth is enabled."
  default     = false
}

variable "access_policies" {
  type        = string
  description = "IAM policy document specifying the access policies for the domain."
}

variable "create_timeout" {
  type        = string
  description = "How much time to wait for the creation before timing out."
  default     = "2h"
}

variable "ebs_enabled" {
  type        = bool
  description = "Whether EBS volumes are attached to data nodes in the domain."
  default     = true
}

variable "ebs_iops" {
  type        = number
  description = "Baseline input/output (I/O) performance of EBS volumes attached to data nodes. Applicable only for the GP3 and Provisioned IOPS EBS volume types."
}

variable "ebs_throughput" {
  type        = number
  description = "(Required if `ebs_volume_type` is set to gp3) Specifies the throughput (in MiB/s) of the EBS volumes attached to data nodes. Applicable only for the gp3 volume type."
}

variable "ebs_volume_type" {
  type        = string
  default     = "gp3"
  description = "Type of EBS volumes attached to data nodes."
}

variable "ebs_volume_size" {
  type        = number
  description = "Size of EBS volumes attached to data nodes."
  required    = true
  default     = 64
}

variable "enable_access_policy" {
  type        = bool
  default     = true
  description = "Determines whether an access policy will be applied to the domain"
}

variable "auto_tune_options" {
  type        = any
  description = "Configuration block for the Auto-Tune options of the domain"
  default     = { "desired_state" : "ENABLED", "rollback_on_disable" : "NO_ROLLBACK" }
}

variable "domain_endpoint_options" {
  type        = any
  description = "Configuration block for domain endpoint HTTP(S) related options"
  default     = { "enforce_https" : true, "tls_security_policy" : "Policy-Min-TLS-1-2-2019-07" }
}

variable "ip_address_type" {
  type        = string
  description = "The IP address type for the endpoint. Valid values are ipv4 and dualstack"
}

variable "off_peak_window_options" {
  type        = any
  description = "Configuration to add Off Peak update options"
  default     = { "enabled" : true, "off_peak_window" : { "hours" : 7 } }
}

variable "kms_key_delete_window_in_days" {
  type        = number
  description = "The number of days before the KMS key is deleted after being disabled."
  default     = 7
}

variable "kms_key_enable_key_rotation" {
  type        = bool
  description = "Specifies whether automatic key rotation is enabled for the KMS key."
  default     = true
}

variable "kms_key_tags" {
  type        = map(string)
  description = "The tags to associate with the KMS key."
  default     = {}
}
