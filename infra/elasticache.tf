# infra/elasticache.tf

# ----------------------------------------------------
# 1. ElastiCache Subnet Group (ElastiCacheを配置するサブネットのグループ化)
# ----------------------------------------------------
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.project_name}-cache-subnet-group"
  # プライベートサブネット（DBと同じく外部から隔離される場所）を指定
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]

  tags = {
    Name = "${var.project_name}-cache-subnet-group"
  }
}

# ----------------------------------------------------
# 2. ElastiCache Cluster (Redis)
# ----------------------------------------------------
resource "aws_elasticache_cluster" "main" {
  cluster_id           = "${var.project_name}-cache"
  engine               = "redis"
  engine_version       = "7.1" # 最新の安定版を指定
  node_type            = "cache.t3.micro" # 開発・デモ用インスタンス
  num_cache_nodes      = 1 # 単一ノード構成 (開発環境のため)
  port                 = 6379 # Redisの標準ポート
  
  # ネットワーク設定
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  # セキュリティグループ（ECS Fargateからのみアクセス許可）
  security_group_ids   = [aws_security_group.redis.id]
  
  # バックアップ設定
  snapshot_retention_limit = 7
  
  tags = {
    Name = "${var.project_name}-cache"
  }
}