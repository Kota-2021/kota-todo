# infra/alb.tf

# ----------------------------------------------------
# 1. Application Load Balancer (ALB) - WebSocket対応
# ----------------------------------------------------
resource "aws_lb" "main" {
  name               = "${var.project_name}-alb"
  internal           = false # インターネット向けALB
  load_balancer_type = "application"
  
  # パブリックサブネットに配置（インターネットからアクセス可能）
  subnets = [
    aws_subnet.public_a.id,
    aws_subnet.public_b.id
  ]

  # セキュリティグループ（既存のALB用SGを使用）
  security_groups = [aws_security_group.alb.id]

  # WebSocket対応: アイドルタイムアウトを3600秒（最大値）に設定
  # デフォルト60秒ではWebSocket接続が切断されるため、長時間接続を維持するために必要
  idle_timeout = 3600

  # 削除保護（本番環境では有効化を推奨）
  # 開発環境ではdeployとdestroyを繰り返し行うため、削除保護を無効化する。
  enable_deletion_protection = false

  tags = {
    Name = "${var.project_name}-alb"
  }
}

# ----------------------------------------------------
# 2. ターゲットグループ (WebSocket対応)
# ----------------------------------------------------
resource "aws_lb_target_group" "main" {
  name     = "${var.project_name}-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  # Fargateではターゲットタイプをipに設定
  target_type = "ip"

  # ヘルスチェック設定
  health_check {
    enabled             = true
    healthy_threshold   = 2   # 正常と判定するまでの連続成功回数
    unhealthy_threshold = 2   # 異常と判定するまでの連続失敗回数
    timeout             = 5   # タイムアウト（秒）
    interval            = 30  # ヘルスチェック間隔（秒）
    path                = "/health" # ヘルスチェックパス（アプリケーションに実装が必要）
    protocol            = "HTTP"
    matcher             = "200" # 正常と判定するHTTPステータスコード
  }

  # WebSocket対応: 登録解除遅延（デフォルト300秒）
  # タスクの停止時に既存接続を維持する時間
  # WebSocket接続を安全に終了させるために重要
  deregistration_delay = 300

  # WebSocket対応: セッションアフィニティ（スティッキーセッション）を有効化
  # 同じクライアントからのリクエストを同じタスクにルーティング
  # WebSocket接続の一貫性を保つために必要
  stickiness {
    enabled         = true
    type            = "lb_cookie"
    cookie_duration = 86400 # 24時間（秒）
  }

  tags = {
    Name = "${var.project_name}-target-group"
  }
}

# ----------------------------------------------------
# 3. ALBリスナー: HTTP (ポート80) - HTTPSへリダイレクト
# ----------------------------------------------------
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# ----------------------------------------------------
# 4. ALBリスナー: HTTPS (ポート443) - WebSocket対応
# ----------------------------------------------------
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  # ACM証明書のARNを変数から取得
  certificate_arn = var.acm_certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}