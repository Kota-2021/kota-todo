provider "aws" {
  region = "ap-northeast-1"

  # 全リソースに自動付与される共通タグ（コスト管理などで便利です）
  default_tags {
    tags = {
    Project   = var.project_name
    ManagedBy = "Terraform"
    }
  }
}