
variable "cluster_name" {
  type        = string
  description = "The name of the ROSA cluster to create"
  default     = "my-ocp-cluster"
}

variable "openshift_version" {
  type        = string
  description = "The version of ROSA to be deployed"
  # TODO renovate latest version
  default = "4.14.21"
  validation {
    condition     = can(regex("^[0-9]*[0-9]+.[0-9]*[0-9]+.[0-9]*[0-9]+$", var.openshift_version))
    error_message = "openshift_version must be with structure <major>.<minor>.<patch> (for example 4.13.6)."
  }
}

variable "replicas" {
  type        = string
  description = "The number of computer nodes to create. Must be a minimum of 2 for a single-AZ cluster, 3 for multi-AZ."
  default     = "2"
}

variable "compute_node_instance_type" {
  type        = string
  description = "The EC2 instance type to use for compute nodes"
  default     = "m5.xlarge"
}

variable "host_prefix" {
  type        = string
  description = "The subnet mask to assign to each compute node in the cluster"
  default     = "23"
}

variable "offline_access_token" {
  type        = string
  description = "The Red Hat OCM API access token for your account"
}

variable "availability_zones_count" {
  type        = number
  description = "The number of availability zones to use for the cluster (minimum 2)"
  default     = 2
}

variable "vpc_cidr_block" {
  type        = string
  description = "value of the CIDR block to use for the VPC"
  default     = "10.66.0.0/16"
}

variable "htpasswd_username" {
  type        = string
  description = "htpasswd username"
  default     = "kubeadmin"
}

variable "htpasswd_password" {
  type        = string
  description = "htpasswd password"
  sensitive   = true
}

variable "url" {
  type        = string
  description = "Provide OCM environment by setting a value to url"
  default     = "https://api.openshift.com"
}
