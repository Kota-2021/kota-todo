# infra/ecr.tf

# ----------------------------------------------------
# 1. ECRリポジトリ (Dockerイメージの保存先)
# ----------------------------------------------------
resource "aws_ecr_repository" "main" {
  name                 = var.project_name
  image_tag_mutability = "IMMUTABLE"

  # 脆弱性スキャンを push 時に実行
  image_scanning_configuration {
    scan_on_push = true
  }

  # AWS 管理キーによる暗号化
  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name = "${var.project_name}-ecr"
  }
}

# ----------------------------------------------------
# 2. ECR ライフサイクルポリシー（古いイメージの自動削除）
# ----------------------------------------------------
resource "aws_ecr_lifecycle_policy" "main" {
  repository = aws_ecr_repository.main.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 production images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 10
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images older than 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}
