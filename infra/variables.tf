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

// ECSタスクのCPU設定（Fargate互換の値）
variable "ecs_task_cpu" {
  description = "CPU units for ECS task (Fargate compatible: 256, 512, 1024, etc.)"
  type        = string
  default     = "256" // 0.25 vCPU (最小値)
}

// ECSタスクのメモリ設定（Fargate互換の値）
variable "ecs_task_memory" {
  description = "Memory for ECS task in MB (Fargate compatible: 512, 1024, 2048, etc.)"
  type        = string
  default     = "512" // 512 MB (最小値)
}

// ECSサービスの希望タスク数
variable "ecs_service_desired_count" {
  description = "Desired number of ECS tasks to run"
  type        = number
  default     = 1
}

// ECSタスク定義のイメージURI
variable "app_image_uri" {
  type        = string
  nullable    = true
  description = "ECRにプッシュされたイメージのURI"
}