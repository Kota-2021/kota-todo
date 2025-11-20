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