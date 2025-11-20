# infra/rds.tf

# ----------------------------------------------------
# 1. RDS DB Subnet Group (DBを配置するサブネットのグループ化)
# ----------------------------------------------------
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  # プライベートサブネット（DBが外部から隔離される場所）を指定
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id] 

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

# ----------------------------------------------------
# 2. RDS Instance (PostgreSQL)
# ----------------------------------------------------
resource "aws_db_instance" "main" {
  # DBの識別子
  identifier = "${var.project_name}-db"

  # エンジンとバージョン
  engine         = "postgres"
  engine_version = "17.6" # 安定版を指定

  # インスタンスタイプ (開発・デモ用 t3.micro)
  # コストを考慮し、t3.microを使用します。
  instance_class = "db.t3.micro" 
  
  # 接続情報
  username = var.db_username
  # パスワードは機密情報としてvariables.tfから取得
  password = var.db_password 

  # ストレージ設定
  allocated_storage    = 20 # 20GB (最小サイズ)
  max_allocated_storage = 100 # 自動拡張の上限
  storage_type         = "gp2"
  storage_encrypted    = true # 暗号化 (セキュリティ要件)

  # ネットワーク設定
  db_subnet_group_name = aws_db_subnet_group.main.name
  # セキュリティグループ（ECSとBastionからのみアクセス許可）
  vpc_security_group_ids = [aws_security_group.rds.id] 
  
  # DBをパブリックに公開しない (プライベートサブネット内のみアクセス可)
  publicly_accessible = false 

  # バックアップ設定
  backup_retention_period = 7 # 7日間保持
  # 「使うときだけ起動」運用のため、最終スナップショットはスキップ
  skip_final_snapshot     = true 

  # RDSの削除保護を無効化（手動での削除を容易にするため）
  deletion_protection = false 
  
  # タグ
  tags = {
    Name = "${var.project_name}-db"
  }
}