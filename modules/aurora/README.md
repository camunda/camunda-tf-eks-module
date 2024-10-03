# RDS Aurora Module

Terraform module which creates Aurora RDS resources with an opinionated configuration targeting Camunda 8.

## Usage

Following is a simple example configuration and should be adjusted as required.

See [inputs](#inputs) for further configuration options and how they affect the RDS creation.

```hcl
module "postgresql" {
  source                     = "github.com/camunda/camunda-tf-eks-module/modules/aurora"
  engine_version             = "15.4"
  auto_minor_version_upgrade = false
  cluster_name               = "cluster-name-postgresql"

  username         = "username"
  password         = "password"
  vpc_id           = module.eks_cluster.vpc_id
  subnet_ids       = module.eks_cluster.private_subnet_ids
  cidr_blocks      = concat(module.eks_cluster.private_vpc_cidr_blocks, module.eks_cluster.public_vpc_cidr_blocks)
  instance_class   = "db.t3.medium"
  iam_auth_enabled = true

  depends_on = [module.eks_cluster]
}
```

<!-- BEGIN_TF_DOCS -->
## Modules

No modules.
## Resources

| Name | Type |
|------|------|
| [aws_db_subnet_group.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_subnet_group) | resource |
| [aws_iam_policy.aurora_access_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.aurora_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.attach_aurora_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_kms_key.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_rds_cluster.aurora_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/rds_cluster) | resource |
| [aws_rds_cluster_instance.aurora_instance](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/rds_cluster_instance) | resource |
| [aws_security_group.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group_rule.allow_egress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
| [aws_security_group_rule.allow_ingress](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule) | resource |
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_auto_minor_version_upgrade"></a> [auto\_minor\_version\_upgrade](#input\_auto\_minor\_version\_upgrade) | If true, minor engine upgrades will be applied automatically to the DB instance during the maintenance window | `bool` | `true` | no |
| <a name="input_availability_zones"></a> [availability\_zones](#input\_availability\_zones) | Array of availability zones to use for the Aurora cluster | `list(string)` | <pre>[<br/>  "eu-central-1a",<br/>  "eu-central-1b",<br/>  "eu-central-1c"<br/>]</pre> | no |
| <a name="input_ca_cert_identifier"></a> [ca\_cert\_identifier](#input\_ca\_cert\_identifier) | Specifies the identifier of the CA certificate for the DB instance | `string` | `"rds-ca-rsa2048-g1"` | no |
| <a name="input_cidr_blocks"></a> [cidr\_blocks](#input\_cidr\_blocks) | The CIDR blocks to allow acces from and to. | `list(string)` | n/a | yes |
| <a name="input_cluster_name"></a> [cluster\_name](#input\_cluster\_name) | Name of the cluster, also used to prefix dependent resources. Format: /[[:lower:][:digit:]-]/ | `any` | n/a | yes |
| <a name="input_default_database_name"></a> [default\_database\_name](#input\_default\_database\_name) | The name for the automatically created database on cluster creation. | `string` | `"camunda"` | no |
| <a name="input_engine"></a> [engine](#input\_engine) | The engine type e.g. aurora, aurora-mysql, aurora-postgresql, ... | `string` | `"aurora-postgresql"` | no |
| <a name="input_engine_version"></a> [engine\_version](#input\_engine\_version) | The DB engine version for Postgres to use. | `string` | `"15.4"` | no |
| <a name="input_iam_aurora_access_policy"></a> [iam\_aurora\_access\_policy](#input\_iam\_aurora\_access\_policy) | Access policy for Aurora allowing access | `string` | `"            {\n              \"Version\": \"2012-10-17\",\n              \"Statement\": [\n                {\n                  \"Effect\": \"Allow\",\n                  \"Action\": [\n                    \"rds-db:connect\"\n                  ],\n                  \"Resource\": \"arn:aws:rds-db:<YOUR-REGION>:<YOUR-ACCOUNT-ID>:dbuser:<YOUR-CLUSTER-NAME>/<YOUR-DB-USER-NAME>\"\n                }\n              ]\n            }\n\n"` | no |
| <a name="input_iam_aurora_role_name"></a> [iam\_aurora\_role\_name](#input\_iam\_aurora\_role\_name) | Name of the AuroraRole IAM role | `string` | `"AuroraRole"` | no |
| <a name="input_iam_auth_enabled"></a> [iam\_auth\_enabled](#input\_iam\_auth\_enabled) | Determines whether IAM auth should be activated for IRSA usage | `bool` | `false` | no |
| <a name="input_iam_create_aurora_role"></a> [iam\_create\_aurora\_role](#input\_iam\_create\_aurora\_role) | Flag to determine if the Aurora IAM role should be created, if true, this module will create a role. Please ensure that iam\_auth\_enabled is set to `true` | `bool` | `true` | no |
| <a name="input_iam_role_trust_policy"></a> [iam\_role\_trust\_policy](#input\_iam\_role\_trust\_policy) | Assume role trust policy for Aurora role | `string` | `"          {\n            \"Version\": \"2012-10-17\",\n            \"Statement\": [\n              {\n                \"Effect\": \"Allow\",\n                \"Principal\": {\n                  \"Federated\": \"arn:aws:iam::<YOUR-ACCOUNT-ID>:oidc-provider/oidc.eks.<YOUR-REGION>.amazonaws.com/id/<YOUR-OIDC-ID>\"\n                },\n                \"Action\": \"sts:AssumeRoleWithWebIdentity\",\n                \"Condition\": {\n                  \"StringEquals\": {\n                    \"oidc.eks.<YOUR-REGION>.amazonaws.com/id/<YOUR-OIDC-PROVIDER-ID>:sub\": \"system:serviceaccount:<YOUR-NAMESPACE>:<YOUR-SA-NAME>\"\n                  }\n                }\n              }\n            ]\n          }\n\n"` | no |
| <a name="input_iam_roles"></a> [iam\_roles](#input\_iam\_roles) | Allows propagating additional IAM roles to the Aurora cluster to allow e.g. access to S3 | `list(string)` | `[]` | no |
| <a name="input_instance_class"></a> [instance\_class](#input\_instance\_class) | The instance type of the Aurora instances | `string` | `"db.t3.medium"` | no |
| <a name="input_num_instances"></a> [num\_instances](#input\_num\_instances) | Number of instances | `string` | `"1"` | no |
| <a name="input_password"></a> [password](#input\_password) | The password for the postgres admin user. Important: secret value! | `string` | n/a | yes |
| <a name="input_subnet_ids"></a> [subnet\_ids](#input\_subnet\_ids) | The subnet IDs to create the cluster in. For easier usage we are passing through the subnet IDs from the AWS EKS Cluster module. | `list(string)` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags to add to the resources | `map` | `{}` | no |
| <a name="input_username"></a> [username](#input\_username) | The username for the postgres admin user. Important: secret value! | `string` | n/a | yes |
| <a name="input_vpc_id"></a> [vpc\_id](#input\_vpc\_id) | The VPC ID to create the cluster in. For easier usage we are passing through the VPC ID from the AWS EKS Cluster module. | `any` | n/a | yes |
## Outputs

| Name | Description |
|------|-------------|
| <a name="output_aurora_endpoint"></a> [aurora\_endpoint](#output\_aurora\_endpoint) | The endpoint of the Aurora cluster |
| <a name="output_aurora_policy_arn"></a> [aurora\_policy\_arn](#output\_aurora\_policy\_arn) | The ARN of the aurora access policy |
| <a name="output_aurora_role_arn"></a> [aurora\_role\_arn](#output\_aurora\_role\_arn) | The ARN of the aurora IAM role |
| <a name="output_aurora_role_name"></a> [aurora\_role\_name](#output\_aurora\_role\_name) | The name of the aurora IAM role |
<!-- END_TF_DOCS -->
