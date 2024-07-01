variable "region" {
  type        = string
  description = "The region where the cluster and relevant resources should be deployed in"
}

variable "name" {
  type        = string
  description = "Name being used for relevant resources - including EKS cluster name"
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version to be used by EKS"
  default     = "1.28"
}

variable "np_min_node_count" {
  type        = number
  description = "Minimum number of nodes for the default node pool"
  default     = 1
}

variable "np_max_node_count" {
  type        = number
  description = "Maximum number of nodes for the default node pool"
  default     = 10
}

variable "np_desired_node_count" {
  type        = number
  description = "Actual number of nodes for the default node pool. Min-Max will be used for autoscaling"
  default     = 4
}

variable "np_instance_types" {
  type        = list(string)
  description = "Allow passing a list of instance types for the auto scaler to select from when scaling the default node pool"
  default     = ["m6i.xlarge"]
}

variable "np_disk_size" {
  type        = number
  description = "Disk size of the nodes on the default node pool"
  default     = 20
}

variable "np_ami_type" {
  description = "Amazon Machine Image"
  type        = string
  default     = "AL2_x86_64"
}

variable "np_capacity_type" {
  type        = string
  default     = "ON_DEMAND"
  description = "Allows setting the capacity type to ON_DEMAND or SPOT to determine stable nodes"
}

variable "cluster_service_ipv4_cidr" {
  description = "The CIDR block to assign Kubernetes service IP addresses from. Between /24 and /12."
  type        = string
}

variable "cluster_node_ipv4_cidr" {
  description = "The CIDR block for public and private subnets of loadbalancers and nodes. Between /28 and /16."
  type        = string
}
