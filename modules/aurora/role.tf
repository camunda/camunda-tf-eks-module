// Convert the list to a map by using the role_name as the key
locals {
  roles_map = { for role in var.iam_roles_with_policies : role.role_name => role }
}

// IAM Role for Aurora
resource "aws_iam_role" "roles" {
  for_each = local.roles_map

  name               = each.key
  assume_role_policy = each.value.trust_policy
}

// IAM Policy for Aurora Access
resource "aws_iam_policy" "access_policies" {
  for_each = local.roles_map

  name        = "${each.key}-access-policy"
  description = "Access policy for ${each.key}"

  policy = each.value.access_policy
}

// Attach the policy to the role
resource "aws_iam_role_policy_attachment" "attach_policies" {
  for_each = local.roles_map

  role       = aws_iam_role.roles[each.key].name
  policy_arn = aws_iam_policy.access_policies[each.key].arn
}
