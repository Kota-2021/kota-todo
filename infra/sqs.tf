# ----------------------------------------------------
# 1. SQS Standard Queue (標準キュー)
# ----------------------------------------------------
resource "aws_sqs_queue" "main" {
  name                      = "${var.project_name}-main-queue"
  delay_seconds             = 0    # メッセージの遅延配信時間 (秒)
  max_message_size          = 262144 # 最大メッセージサイズ (バイト、デフォルト256KB)
  message_retention_seconds = 86400  # メッセージ保持期間 (秒、1日)
  visibility_timeout_seconds = 30   # 可視性タイムアウト (秒)

  tags = {
    Name = "${var.project_name}-main-queue"
  }
}

# ----------------------------------------------------
# 2. SQS Dead Letter Queue (DLQ)
# ----------------------------------------------------
# メッセージ処理が複数回失敗した場合に、メッセージを移動させるキュー
resource "aws_sqs_queue" "dlq" {
  name                      = "${var.project_name}-main-queue-dlq"
  message_retention_seconds = 1209600 # 保持期間 (14日間)

  tags = {
    Name = "${var.project_name}-main-queue-dlq"
  }
}

# ----------------------------------------------------
# 3. DLQをメインキューに紐づけ
# ----------------------------------------------------
resource "aws_sqs_queue_redrive_policy" "main_queue_redrive_policy" {
  queue_url = aws_sqs_queue.main.id
  # 処理失敗が5回続いたらDLQへ移動させる設定
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 5
  })
}

# 20251211byKota 以下を追加する

# 1. DLQ (Dead Letter Queue) の定義
resource "aws_sqs_queue" "task_notification_dlq" {
  # 修正 1: 命名規則の統一
  name                        = "${var.project_name}-task-notification-dlq"
  message_retention_seconds   = 1209600 # 14 days (デフォルト)
  tags = {
    # 修正 1: タグにも反映
    Name        = "${var.project_name}-task-notification-dlq"
    Environment = "dev"
  }
}

# 2. メインキューの定義
resource "aws_sqs_queue" "task_notification_queue" {

  name                        = "${var.project_name}-task-notification-queue"
  
  # 可視性タイムアウト (VisibilityTimeout) - 計画通り60秒
  visibility_timeout_seconds  = 60 

  # メッセージ保持期間を明示（仮に1日を設定）
  message_retention_seconds = 86400 # 1日 (通知メッセージが古くなっても不要な場合)

  #  ロングポーリングの有効化 (最大20秒)
  # これにより、メッセージ受信コストを削減し、レイテンシを改善
  receive_wait_time_seconds   = 20
  
  # 1-3c & 1-3d. Redrive Policy の設定（DLQと最大再試行回数の関連付け）
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.task_notification_dlq.arn
    # 1-3c. 最大再試行回数 (MaxReceiveCount) - 3回失敗したらDLQへ
    maxReceiveCount     = 3 
  })

  tags = {
    Name        = "${var.project_name}-task-notification-queue"
    Environment = "dev"
  }
}