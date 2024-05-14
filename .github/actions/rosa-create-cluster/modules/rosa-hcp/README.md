# rosa-hcp

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_htpasswd_idp"></a> [htpasswd\_idp](#module\_htpasswd\_idp) | terraform-redhat/rosa-hcp/rhcs//modules/idp | 1.6.2-prerelease.1 |
| <a name="module_rosa_hcp"></a> [rosa\_hcp](#module\_rosa\_hcp) | terraform-redhat/rosa-hcp/rhcs | 1.6.2-prerelease.1 |
| <a name="module_vpc"></a> [vpc](#module\_vpc) | terraform-redhat/rosa-hcp/rhcs//modules/vpc | 1.6.2-prerelease.1 |
## Resources

| Name | Type |
|------|------|
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_availability_zones_count"></a> [availability\_zones\_count](#input\_availability\_zones\_count) | The number of availability zones to use for the cluster (minimum 2) | `number` | `2` | no |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | The name of the ROSA cluster to create | `string` | `"my-ocp-cluster"` | no |
| <a name="input_compute_node_instance_type"></a> [compute\_node\_instance\_type](#input\_compute\_node\_instance\_type) | The EC2 instance type to use for compute nodes | `string` | `"m5.xlarge"` | no |
| <a name="input_host_prefix"></a> [host\_prefix](#input\_host\_prefix) | The subnet mask to assign to each compute node in the cluster | `string` | `"23"` | no |
| <a name="input_htpasswd_password"></a> [htpasswd\_password](#input\_htpasswd\_password) | htpasswd password | `string` | n/a | yes |
| <a name="input_htpasswd_username"></a> [htpasswd\_username](#input\_htpasswd\_username) | htpasswd username | `string` | `"kubeadmin"` | no |
| <a name="input_offline_access_token"></a> [offline\_access\_token](#input\_offline\_access\_token) | The Red Hat OCM API access token for your account | `string` | n/a | yes |
| <a name="input_openshift_version"></a> [openshift\_version](#input\_openshift\_version) | The version of ROSA to be deployed | `string` | `"4.14.21"` | no |
| <a name="input_replicas"></a> [replicas](#input\_replicas) | The number of computer nodes to create. Must be a minimum of 2 for a single-AZ cluster, 3 for multi-AZ. | `string` | `"2"` | no |
| <a name="input_url"></a> [url](#input\_url) | Provide OCM environment by setting a value to url | `string` | `"https://api.openshift.com"` | no |
| <a name="input_vpc_cidr_block"></a> [vpc\_cidr\_block](#input\_vpc\_cidr\_block) | value of the CIDR block to use for the VPC | `string` | `"10.66.0.0/16"` | no |
## Outputs

| Name | Description |
|------|-------------|
| <a name="output_all_subnets"></a> [all\_subnets](#output\_all\_subnets) | For use as '--subnet-ids' parameter in rosa command |
| <a name="output_cluster_id"></a> [cluster\_id](#output\_cluster\_id) | n/a |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | n/a |
| <a name="output_public_subnet_ids"></a> [public\_subnet\_ids](#output\_public\_subnet\_ids) | n/a |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
