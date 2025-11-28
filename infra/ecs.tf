# infra/ecs.tf

# ----------------------------------------------------
# 1. CloudWatch Logs ロググループ (ECSコンテナのログ保存先)
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group
resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = 7 # ログ保持期間（7日間）

  tags = {
    Name = "${var.project_name}-ecs-logs"
  }
}

# ----------------------------------------------------
# 2. ECS Fargateクラスター
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"

  # Container Insights を有効化（メトリクスとログの可視化）
  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  # 設定ブロック（タグなど）
  tags = {
    Name = "${var.project_name}-cluster"
  }
}

# ----------------------------------------------------
# 3. IAMロール: ECSタスク実行ロール (ECSエージェントが使用)
# ----------------------------------------------------
# ECSエージェントがECRからイメージをプルし、CloudWatch Logsにログを送信するためのロール
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role
resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.project_name}-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-ecs-task-execution-role"
  }
}

# ECRからイメージをプルする権限
# 以下のURLを参照
resource "aws_iam_role_policy_attachment" "ecs_task_execution_ecr" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# CloudWatch Logsにログを送信する権限（上記ポリシーに含まれているが、明示的に追加可能）
# AmazonECSTaskExecutionRolePolicyに既に含まれているため、追加不要

# ----------------------------------------------------
# 4. IAMロール: ECSタスクロール (アプリケーションが使用)
# ----------------------------------------------------
# アプリケーションがAWSサービス（RDS、SQS、ElastiCacheなど）にアクセスするためのロール
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role
resource "aws_iam_role" "ecs_task" {
  name = "${var.project_name}-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      },
      # ★ Secrets Manager アクセス権限を追記 ★
      {
        Sid    = "SecretsManagerAccess"
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          # RDSパスワードシークレットのARNを指定
          "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.project_name}/db-password-*"
        ]
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-ecs-task-role"
  }
}

# SQSへのアクセス権限（必要に応じて追加）
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy
resource "aws_iam_role_policy" "ecs_task_sqs" {
  name = "${var.project_name}-ecs-task-sqs-policy"
  role = aws_iam_role.ecs_task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          aws_sqs_queue.main.arn,
          aws_sqs_queue.dlq.arn
        ]
      }
    ]
  })
}

# データソース: AWSアカウントIDを取得（ECR URI構築用）
data "aws_caller_identity" "current" {}

# データソース: AWSリージョン情報を取得
data "aws_region" "current" {}

# ----------------------------------------------------
# 5. ECSタスク定義 (Fargate)
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_task_definition
resource "aws_ecs_task_definition" "main" {
  family                   = "${var.project_name}-task"
  network_mode             = "awsvpc" # Fargateでは必須
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.ecs_task_cpu
  memory                   = var.ecs_task_memory
  
  # タスク実行ロール（ECR、CloudWatch Logsへのアクセス用）
  execution_role_arn = aws_iam_role.ecs_task_execution.arn
  # タスクロール（アプリケーションがAWSサービスにアクセスする用）
  task_role_arn = aws_iam_role.ecs_task.arn

  # コンテナ定義
  container_definitions = jsonencode([
    {
      name  = "${var.project_name}-api"
      image     = var.app_image_uri,
      # image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_repository.main.name}:latest"
      
      # ポートマッピング
      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      # 環境変数
      environment = [
        {
          name  = "DB_HOST"
          value = aws_db_instance.main.address
        },
        {
          name  = "DB_PORT"
          value = tostring(aws_db_instance.main.port)
        },
        {
          name  = "DB_USER"
          value = aws_db_instance.main.username
        },
        {
          name  = "DB_NAME"
          value = "portfolio_db" # RDS作成時に指定したDB名
        },
        {
          name  = "DB_SSLMODE"
          value = "require" # 本番環境ではSSL必須
        },
        {
          name  = "REDIS_HOST"
          value = aws_elasticache_cluster.main.cache_nodes[0].address
        },
        {
          name  = "REDIS_PORT"
          value = tostring(aws_elasticache_cluster.main.port)
        },
        {
          name  = "SQS_QUEUE_URL"
          value = aws_sqs_queue.main.url
        }
        # 注意: DB_PASSWORDは機密情報のため、本番環境では必ずSecrets Managerを使用すること
        # 開発環境では環境変数として設定（セキュリティリスクがあるため本番では非推奨）
        # 本番環境では以下のsecretsブロックを使用すること
      ]

      # シークレット（機密情報は環境変数ではなくシークレットとして管理することを推奨）
      # 本番環境ではSecrets ManagerやParameter Storeを使用することを推奨
      # 開発環境では上記の環境変数として設定（セキュリティリスクがあるため本番では非推奨）
      secrets = [
        {
          name      = "DB_PASSWORD"
          valueFrom = "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.project_name}/db-password:password::"
        }
      ]

      # ログ設定（CloudWatch Logs）
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "ecs"
        }
      }

      # ヘルスチェック（オプション、アプリケーションにヘルスチェックエンドポイントがある場合）
      # healthCheck = {
      #   command     = ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
      #   interval    = 30
      #   timeout     = 5
      #   retries     = 3
      #   startPeriod = 60
      # }

      # 必須設定
      essential = true
    }
  ])

  tags = {
    Name = "${var.project_name}-task-definition"
  }
}

# ----------------------------------------------------
# 6. ECSサービス (Fargate)
# ----------------------------------------------------
# 以下のURLを参照
# https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service
resource "aws_ecs_service" "main" {
  name            = "${var.project_name}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.ecs_service_desired_count

  launch_type = "FARGATE"

  # ネットワーク設定
  network_configuration {
    subnets = [
      aws_subnet.private_a.id,
      aws_subnet.private_b.id
    ]
    security_groups = [aws_security_group.ecs_fargate.id]
    # パブリックIPは不要（プライベートサブネット内で実行）
    assign_public_ip = false
  }

  # ロードバランサー設定
  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = "${var.project_name}-api"
    container_port   = 8080
  }

  # デプロイメント設定（ローリングデプロイメント）
  # maximum_percent: 最大200%までスケールアップ可能
  # minimum_healthy_percent: 最小100%を維持（ゼロダウンタイムデプロイ）
  # デフォルト値を使用（maximum_percent=200, minimum_healthy_percent=100）

  # ヘルスチェック設定（ECSサービスレベル）
  health_check_grace_period_seconds = 60 # ヘルスチェック開始までの猶予期間（秒）

  # デプロイメントサーキットブレーカー（別ブロック）
  deployment_circuit_breaker {
    enable   = true  # デプロイメントサーキットブレーカーを有効化
    rollback = true  # 失敗時に自動ロールバック
  }

  # タスク定義の変更を検知して自動更新
  # 新しいタスク定義が作成された場合、サービスを手動で更新する必要がある
  # または、CI/CDパイプラインで自動更新

  tags = {
    Name = "${var.project_name}-service"
  }

  # ターゲットグループが作成されてからサービスを作成
  depends_on = [aws_lb_target_group.main]
}

