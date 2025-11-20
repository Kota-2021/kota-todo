terraform {
  // Terraform本体のバージョン制約
  required_version = ">= 1.13.5"

  // プロバイダ（AWS）のバージョン制約
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      // 今回のコード作成時点での安定版バージョンを指定
      version = "~> 6.21"
    }
  }
}