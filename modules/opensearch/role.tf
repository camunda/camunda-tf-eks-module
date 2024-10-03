// IAM Role for OpenSearch
resource "aws_iam_role" "opensearch" {
  count = var.iam_create_opensearch_role ? 1 : 0

  name               = var.iam_opensearch_role_name
  assume_role_policy = var.iam_role_trust_policy
}

// IAM Policy for OpenSearch Access
resource "aws_iam_policy" "opensearch_access_policy" {
  count = var.iam_create_opensearch_role ? 1 : 0

  name        = "${var.iam_opensearch_role_name}-access-policy"
  description = "Access policy for OpenSearch"

  policy = var.iam_opensearch_access_policy
}

// Attach the policy to the role
resource "aws_iam_role_policy_attachment" "attach_opensearch_policy" {
  count = var.iam_create_opensearch_role ? 1 : 0

  role       = aws_iam_role.opensearch[0].name
  policy_arn = aws_iam_policy.opensearch_access_policy[0].arn
}
