output "public_subnet_ids" {
  value = join(",", module.vpc.public_subnets)
}

output "private_subnet_ids" {
  value = join(",", module.vpc.private_subnets)
}

output "all_subnets" {
  value       = join(",", concat(module.vpc.private_subnets, module.vpc.public_subnets))
  description = "For use as '--subnet-ids' parameter in rosa command"
}

output "cluster_id" {
  value = module.rosa_hcp.cluster_id
}
