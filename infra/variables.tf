// プロジェクト名
variable "project_name" {
  description = "Project name to be used for resource tagging."
  type        = string
  default     = "my-portfolio-2025"
}

// VPCのCIDRブロック
variable "vpc_cidr" {
  description = "The CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

// RDSのマスターパスワード（機密情報のため、ここでは仮の値。実際は外部から渡す）
variable "db_password" {
  description = "Master password for the RDS instance."
  type        = string
  sensitive   = true // Stateファイルに平文で保存されないように設定
}

// VPCのリージョン
variable "vpc_region" {
  description = "Region for the AWS resources."
  type        = string
  default     = "ap-northeast-1"
}