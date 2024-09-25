resource "aws_opensearch_domain" "opensearch_cluster" {

  tags = var.tags

  domain_name    = var.domain_name
  engine_version = var.engine_version

  ip_address_type = var.ip_address_type

  vpc_options {
    vpc_id             = var.vpc_id
    subnet_ids         = var.subnet_ids
    security_group_ids = var.security_group_ids
    availability_zones = var.availability_zones
  }

  off_peak_window_options = var.off_peak_window_options

  # TODO: integrate logwatch in this component but also in the other for production ready solution

  cluster_config {
    instance_type  = var.instance_type
    instance_count = var.instance_count

    cold_storage_options {
      enabled = var.cold_storage_enabled
    }

    dedicated_master_enabled      = var.dedicated_master_enabled
    dedicated_master_type         = var.dedicated_master_type
    dedicated_master_count        = var.dedicated_master_count
    multi_az_with_standby_enabled = var.multi_az_with_standby_enabled

    warm_enabled = var.warm_enabled
    warm_count   = var.warm_count
    warm_type    = var.warm_type

    zone_awareness_config {
      availability_zone_count = var.zone_awareness_availability_zone_count
    }
    zone_awareness_enabled = var.zone_awareness_enabled
  }

  software_update_options = {
    auto_software_update_enabled = var.auto_software_update_enabled
  }

  advanced_security_options {
    enabled                        = var.advanced_security_enabled
    internal_user_database_enabled = var.advanced_security_internal_user_database_enabled

    master_user_options {
      master_user_name     = var.advanced_security_master_user_name
      master_user_password = var.advanced_security_master_user_password
    }

    anonymous_auth_enabled = var.advanced_security_anonymous_auth_enabled
  }

  encrypt_at_rest {
    enabled    = var.encrypt_at_rest_enabled
    kms_key_id = var.encrypt_at_rest_kms_key_id
  }

  node_to_node_encryption {
    enabled = var.node_to_node_encryption_enabled
  }

  ebs_options {
    ebs_enabled = var.ebs_enabled
    iops        = var.ebs_iops
    volume_size = var.ebs_volume_size
    volume_type = var.ebs_volume_type
    throughput  = var.ebs_throughput
  }

  snapshot_options {
    automated_snapshot_start_hour = var.automated_snapshot_start_hour
  }

  auto_tune_options = var.auto_tune_options

  advanced_options = var.advanced_options

  enable_access_policy = var.enable_access_policy
  access_policies      = var.access_policies

  domain_endpoint_options = var.domain_endpoint_options

  timeouts {
    create = var.create_timeout
  }

}

# TODO: add kms key, security group, subnet, inspire on aurora
