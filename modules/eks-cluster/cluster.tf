/*
The following 2 data resources are used get around the fact that we have to wait
for the EKS cluster to be initialised before we can attempt to authenticate.
*/

data "aws_eks_cluster" "eks" {
  name = module.eks.cluster_name

  # depend on something of the eks module but nothing that would ever change
  # workaround to only pull data on a later stage during initial creation
  depends_on = [
    module.eks.cluster_name
  ]
}

data "aws_eks_cluster_auth" "eks" {
  name = module.eks.cluster_name

  # depend on something of the eks module but nothing that would ever change
  # workaround to only pull data on a later stage during initial creation
  depends_on = [
    module.eks.cluster_name
  ]
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.eks.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.eks.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.eks.token
}

# https://github.com/terraform-aws-modules/terraform-aws-eks
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.15.0"

  cluster_name                    = var.name
  cluster_version                 = var.kubernetes_version
  cluster_service_ipv4_cidr       = var.cluster_service_ipv4_cidr
  cluster_endpoint_private_access = true # private API communication for nodes within the VPC
  cluster_endpoint_public_access  = true # API accessible to engineers

  cluster_addons = {
    coredns = {
      most_recent       = true
      resolve_conflicts = "OVERWRITE"
    }
    kube-proxy = {
      most_recent       = true
      resolve_conflicts = "OVERWRITE"
    }
    vpc-cni = {
      most_recent       = true
      resolve_conflicts = "OVERWRITE"
    }
    aws-ebs-csi-driver = {
      most_recent       = true
      resolve_conflicts = "OVERWRITE"

      before_compute           = true
      service_account_role_arn = module.ebs_cs_role.iam_role_arn
    }
  }

  cluster_encryption_config = {
    provider_key_arn = aws_kms_key.eks.arn
    resources        = ["secrets"]
  }

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  enable_irsa = true # https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/

  # Extend cluster security group rules
  cluster_security_group_additional_rules = {
    egress_nodes_ephemeral_ports_tcp = {
      description                = "To node 1025-65535"
      protocol                   = "tcp"
      from_port                  = 1025
      to_port                    = 65535
      type                       = "egress"
      source_node_security_group = true
    }
  }

  # Extend node-to-node security group rules
  node_security_group_additional_rules = {
    ingress_self_all = {
      description = "Node to node all ports/protocols"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      self        = true
    }
    egress_all = {
      description      = "Node all egress"
      protocol         = "-1"
      from_port        = 0
      to_port          = 0
      type             = "egress"
      cidr_blocks      = ["0.0.0.0/0"]
      ipv6_cidr_blocks = ["::/0"]
    }
  }

  # EKS Managed Node Group(s) default values
  eks_managed_node_group_defaults = {
    ami_type       = var.np_ami_type
    disk_size      = var.np_disk_size
    instance_types = var.np_instance_types
    capacity_type  = var.np_capacity_type

    update_config = {
      max_unavailable = 1
    }

    metadata_options = {
      http_put_response_hop_limit = 1 # related to https://stackoverflow.com/a/73958206, don't allow pods to assume the role of a node
    }

    use_custom_launch_template = false

    min_size     = var.np_min_node_count
    max_size     = var.np_max_node_count
    desired_size = var.np_desired_node_count

  }

  # EKS Managed Node Group definitions
  eks_managed_node_groups = {
    services = {
      labels          = {}
      name            = "services"
      use_name_prefix = false
    }
  }

  # creates separate IAM role
  create_iam_role          = true
  iam_role_description     = "${var.name} - EKS managed IAM role"
  iam_role_name            = "${var.name}-eks-iam-role"
  iam_role_use_name_prefix = false
  iam_role_additional_policies = {
    AmazonEC2ContainerRegistryReadOnly = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  }

  node_security_group_name = "${var.name}-eks-node-sg"

  authentication_mode                      = var.authentication_mode
  access_entries                           = var.access_entries
  enable_cluster_creator_admin_permissions = var.enable_cluster_creator_admin_permissions
}

# gp3 storage class
resource "kubernetes_storage_class_v1" "ebs_sc" {
  metadata {
    name = "ebs-sc"
    annotations = {
      "storageclass.kubernetes.io/is-default-class" = "true"
    }
  }
  storage_provisioner = "ebs.csi.aws.com"
  reclaim_policy      = "Retain"
  parameters = {
    type = "gp3" # starting eks 1.30, gp3 is the default
  }
  volume_binding_mode = "WaitForFirstConsumer"

  depends_on = [
    kubernetes_annotations.default_storageclass
  ]
}
