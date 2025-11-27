# infra/ecr.tf

# ----------------------------------------------------
# 1. ECRリポジトリ (Dockerイメージの保存先)
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecr_repository
resource "aws_ecr_repository" "main" {
  name                 = var.project_name
  image_tag_mutability = "IMMUTABLE" # イメージタグの不変性を有効化

  # イメージスキャン設定（セキュリティ脆弱性の自動検出）
  image_scanning_configuration {
    scan_on_push = true
  }

  # 暗号化設定
  encryption_configuration {
    encryption_type = "AES256" # AWS管理キーによる暗号化
  }

  tags = {
    Name = "${var.project_name}-ecr"
  }
}

# ----------------------------------------------------
# 2. ECRライフサイクルポリシー (古いイメージの自動削除)
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecr_lifecycle_policy
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

