// IAM Role for Aurora
resource "aws_iam_role" "aurora_role" {
  count = var.iam_create_aurora_role ? 1 : 0

  name               = var.iam_aurora_role_name
  assume_role_policy = var.iam_role_trust_policy
}

// IAM Policy for Aurora Access
resource "aws_iam_policy" "aurora_access_policy" {
  count = var.iam_create_aurora_role ? 1 : 0

  name        = "${var.iam_aurora_role_name}-access-policy"
  description = "Access policy for Aurora"

  policy = var.iam_aurora_access_policy
}

// Attach the policy to the role
resource "aws_iam_role_policy_attachment" "attach_aurora_policy" {
  count = var.iam_create_aurora_role ? 1 : 0

  role       = aws_iam_role.aurora_role[0].name
  policy_arn = aws_iam_policy.aurora_access_policy[0].arn
}
