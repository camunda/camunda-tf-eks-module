# AWS EKS Cluster Module

Terraform module which creates AWS EKS (Kubernetes) resources with an opinionated configuration targeting Camunda 8.

## Usage

Following is a simple example configuration and should be adjusted as required.

See [inputs](#inputs) for further configuration options and how they affect the cluster creation.

```hcl
module "eks_cluster" {
  source = "github.com/camunda/camunda-tf-eks-module/modules/eks-cluster"

  region             = "eu-central-1"
  name               = "cluster-name"

  cluster_service_ipv4_cidr = "10.190.0.0/16"
  cluster_node_ipv4_cidr    = "10.192.0.0/16"
}
```

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_cert_manager_role"></a> [cert\_manager\_role](#module\_cert\_manager\_role) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks | 5.34.0 |
| <a name="module_ebs_cs_role"></a> [ebs\_cs\_role](#module\_ebs\_cs\_role) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks | 5.34.0 |
| <a name="module_eks"></a> [eks](#module\_eks) | terraform-aws-modules/eks/aws | 19.21.0 |
| <a name="module_external_dns_role"></a> [external\_dns\_role](#module\_external\_dns\_role) | terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks | 5.34.0 |
| <a name="module_vpc"></a> [vpc](#module\_vpc) | terraform-aws-modules/vpc/aws | 5.5.2 |
## Resources

| Name | Type |
|------|------|
| [aws_iam_policy.cert_manager_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_policy.ebs_sc_access](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_policy.ebs_sc_access_2](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_policy.eks_admin_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_policy.external_dns_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_kms_key.eks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_security_group_rule.cluster_api_to_nodes](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [kubernetes_annotations.default_storageclass](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/annotations) | resource |
| [kubernetes_storage_class_v1.ebs_sc](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/storage_class_v1) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_eks_cluster.eks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/eks_cluster) | data source |
| [aws_eks_cluster_auth.eks](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/eks_cluster_auth) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_aws_auth_roles"></a> [aws\_auth\_roles](#input\_aws\_auth\_roles) | List of objects, allowing to map IAM roles to K8s roles. Further information can be found on aws: https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html | <pre>list(object({<br>    rolearn  = string<br>    username = string<br>    groups   = list(string)<br>  }))</pre> | `[]` | no |
| <a name="input_aws_auth_users"></a> [aws\_auth\_users](#input\_aws\_auth\_users) | List of objects, allowing to map IAM users to K8s roles. Further information can be found on aws: https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html | <pre>list(object({<br>    userarn  = string<br>    username = string<br>    groups   = list(string)<br>  }))</pre> | `[]` | no |
| <a name="input_cluster_node_ipv4_cidr"></a> [cluster\_node\_ipv4\_cidr](#input\_cluster\_node\_ipv4\_cidr) | The CIDR block for public and private subnets of loadbalancers and nodes. Between /28 and /16. | `string` | n/a | yes |
| <a name="input_cluster_service_ipv4_cidr"></a> [cluster\_service\_ipv4\_cidr](#input\_cluster\_service\_ipv4\_cidr) | The CIDR block to assign Kubernetes service IP addresses from. Between /24 and /12. | `string` | n/a | yes |
| <a name="input_kubernetes_version"></a> [kubernetes\_version](#input\_kubernetes\_version) | Kubernetes version to be used by EKS | `string` | `"1.28"` | no |
| <a name="input_name"></a> [name](#input\_name) | Name being used for relevant resources - including EKS cluster name | `string` | n/a | yes |
| <a name="input_np_ami_type"></a> [np\_ami\_type](#input\_np\_ami\_type) | Amazon Machine Image | `string` | `"AL2_x86_64"` | no |
| <a name="input_np_capacity_type"></a> [np\_capacity\_type](#input\_np\_capacity\_type) | Allows setting the capacity type to ON\_DEMAND or SPOT to determine stable nodes | `string` | `"ON_DEMAND"` | no |
| <a name="input_np_desired_node_count"></a> [np\_desired\_node\_count](#input\_np\_desired\_node\_count) | Actual number of nodes for the default node pool. Min-Max will be used for autoscaling | `number` | `4` | no |
| <a name="input_np_disk_size"></a> [np\_disk\_size](#input\_np\_disk\_size) | Disk size of the nodes on the default node pool | `number` | `20` | no |
| <a name="input_np_instance_types"></a> [np\_instance\_types](#input\_np\_instance\_types) | Allow passing a list of instance types for the auto scaler to select from when scaling the default node pool | `list(string)` | <pre>[<br>  "m6i.xlarge"<br>]</pre> | no |
| <a name="input_np_max_node_count"></a> [np\_max\_node\_count](#input\_np\_max\_node\_count) | Maximum number of nodes for the default node pool | `number` | `10` | no |
| <a name="input_np_min_node_count"></a> [np\_min\_node\_count](#input\_np\_min\_node\_count) | Minimum number of nodes for the default node pool | `number` | `1` | no |
| <a name="input_region"></a> [region](#input\_region) | The region where the cluster and relevant resources should be deployed in | `string` | n/a | yes |
## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cert_manager_arn"></a> [cert\_manager\_arn](#output\_cert\_manager\_arn) | Amazon Resource Name of the cert-manager IAM role used for IAM Roles to Service Accounts mappings |
| <a name="output_cluster_endpoint"></a> [cluster\_endpoint](#output\_cluster\_endpoint) | Endpoint for your Kubernetes API server |
| <a name="output_cluster_iam_role_arn"></a> [cluster\_iam\_role\_arn](#output\_cluster\_iam\_role\_arn) | IAM role ARN of the EKS cluster |
| <a name="output_cluster_iam_role_name"></a> [cluster\_iam\_role\_name](#output\_cluster\_iam\_role\_name) | IAM role name of the EKS cluster |
| <a name="output_cluster_primary_security_group_id"></a> [cluster\_primary\_security\_group\_id](#output\_cluster\_primary\_security\_group\_id) | Cluster primary security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication. Referred to as 'Cluster security group' in the EKS console |
| <a name="output_cluster_security_group_arn"></a> [cluster\_security\_group\_arn](#output\_cluster\_security\_group\_arn) | Amazon Resource Name (ARN) of the cluster security group |
| <a name="output_cluster_security_group_id"></a> [cluster\_security\_group\_id](#output\_cluster\_security\_group\_id) | Cluster security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication. Referred to as 'Cluster security group' in the EKS console |
| <a name="output_default_security_group_id"></a> [default\_security\_group\_id](#output\_default\_security\_group\_id) | The ID of the security group created by default on VPC creation |
| <a name="output_ebs_cs_arn"></a> [ebs\_cs\_arn](#output\_ebs\_cs\_arn) | Amazon Resource Name of the ebs-csi IAM role used for IAM Roles to Service Accounts mappings |
| <a name="output_external_dns_arn"></a> [external\_dns\_arn](#output\_external\_dns\_arn) | Amazon Resource Name of the external-dns IAM role used for IAM Roles to Service Accounts mappings |
| <a name="output_private_route_table_ids"></a> [private\_route\_table\_ids](#output\_private\_route\_table\_ids) | The IDs of the private route tables associated with this VPC |
| <a name="output_private_subnet_ids"></a> [private\_subnet\_ids](#output\_private\_subnet\_ids) | Private subnet IDs |
| <a name="output_private_vpc_cidr_blocks"></a> [private\_vpc\_cidr\_blocks](#output\_private\_vpc\_cidr\_blocks) | Private VPC CIDR blocks |
| <a name="output_public_vpc_cidr_blocks"></a> [public\_vpc\_cidr\_blocks](#output\_public\_vpc\_cidr\_blocks) | Public VPC CIDR blocks |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | VPC id of the cluster |
| <a name="output_vpc_main_route_table_id"></a> [vpc\_main\_route\_table\_id](#output\_vpc\_main\_route\_table\_id) | The ID of the main route table associated with this VPC |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
