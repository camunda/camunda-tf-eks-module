output "aurora_endpoint" {
  value = "${aws_rds_cluster.aurora_cluster.endpoint}:${aws_rds_cluster.aurora_cluster.port}"
}
