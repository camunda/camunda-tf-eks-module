# TODO: rethink assuming policy stuff
# resource "aws_iam_role" "eks_admin_role" {
#   name = "${var.name}-eks-admin-role"

#   assume_role_policy = jsonencode({

#   })
# }

# resource "aws_iam_role_policy_attachment" "github_infra_core_admin_eks_access" {
#   policy_arn = aws_iam_policy.eks_admin_policy.policy
#   role       = aws_iam_role.eks_admin_role.name
# }

################################################################################
# IRSA
################################################################################

module "ebs_cs_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "5.30.0"

  role_name = "${var.name}-ebs-cs-role"

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }

  role_policy_arns = {
    policy = aws_iam_policy.ebs_sc_access.arn
    policy = aws_iam_policy.ebs_sc_access_2.arn
  }
}
