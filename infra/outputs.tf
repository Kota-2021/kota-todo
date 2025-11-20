# infra/outputs.tf

# ----------------------------------------------------
# 1. ネットワーク情報
# ----------------------------------------------------
output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.main.id
}

# ----------------------------------------------------
# 2. データベース情報
# ----------------------------------------------------
output "rds_endpoint" {
  description = "The hostname of the RDS instance"
  value       = aws_db_instance.main.address
}
output "rds_port" {
  description = "The port of the RDS instance"
  value       = aws_db_instance.main.port
}

# ----------------------------------------------------
# 3. キャッシュ情報
# ----------------------------------------------------
output "elasticache_endpoint" {
  description = "The hostname of the ElastiCache cluster"
  # ConfigurationEndpointはレプリカセット全体への接続に使用されるエンドポイント
  value       = aws_elasticache_cluster.main.cache_nodes[0].address
}
output "elasticache_port" {
  description = "The port of the ElastiCache cluster"
  value       = aws_elasticache_cluster.main.port
}

# ----------------------------------------------------
# 4. SQS情報
# ----------------------------------------------------
output "sqs_main_queue_url" {
  description = "The URL of the main SQS queue"
  value       = aws_sqs_queue.main.id
}
output "sqs_main_queue_arn" {
  description = "The ARN of the main SQS queue"
  value       = aws_sqs_queue.main.arn
}

# ----------------------------------------------------
# 5. セキュリティグループ情報
# ----------------------------------------------------
output "ecs_fargate_sg_id" {
  description = "The ID of the ECS Fargate security group"
  value       = aws_security_group.ecs_fargate.id
}
output "alb_sg_id" {
  description = "The ID of the ALB security group"
  value       = aws_security_group.alb.id
}