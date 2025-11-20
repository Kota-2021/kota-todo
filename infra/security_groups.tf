// ----------------------------------------------------
// 1. ALB (Application Load Balancer) 用 SG
// ----------------------------------------------------
resource "aws_security_group" "alb" {
  name        = "${var.project_name}-sg-alb"
  vpc_id      =aws_vpc.main.id
  description = "Allows HTTP/HTTPS access from internet."

  // インバウンド：インターネット(0.0.0.0/0)からのHTTP(80)とHTTPS(443)を許可
  ingress {
    description = "HTTP access"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    description = "HTTPS access"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  // アウトバウンド：必要最低限に制限
  
  // 1. ECS Fargateへの接続 (アプリケーションポート: 8080)
  egress {
    description     = "To ECS Fargate Targets"
    from_port       = 8080 // ECSが待ち受けるポート
    to_port         = 8080
    protocol        = "tcp"
    // 宛先をECS FargateのSGに限定
    security_groups = [aws_security_group.ecs_fargate.id] 
  }

  // 2. 外部へのHTTPS通信 (CloudWatchへのメトリクス送信、ALB管理など)
  egress {
    description     = "To Internet/AWS APIs (HTTPS)"
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    cidr_blocks     = ["0.0.0.0/0"]
  }
}


// ----------------------------------------------------
// 2. ECS Fargate (Web API) 用 SG
// ----------------------------------------------------
resource "aws_security_group" "ecs_fargate" {
  name        = "${var.project_name}-sg-ecs-fargate"
  vpc_id      = aws_vpc.main.id
  description = "Allows traffic from ALB to ECS Fargate."

  // インバウンド：ALBのSGからWeb APIのポート（例: 8080）へのアクセスを許可
  ingress {
    description     = "From ALB"
    from_port       = 8080 // Goアプリが待ち受けるポート
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id] // ソースをALBのSGに限定
  }

  // アウトバウンド：必要最低限に制限
  
  // 1. RDSへの接続 (5432)
  egress {
    description     = "To RDS"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.rds.id]
  }

  // 2. Redisへの接続 (6379)
  egress {
    description     = "To Redis"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.redis.id]
  }

  // 3. 外部へのHTTPS通信 (AWS API, ECR Pull, SQSなど)
  egress {
    description     = "To Internet (HTTPS for AWS/Packages)"
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    cidr_blocks     = ["0.0.0.0/0"] // NAT Gateway経由で外へ
  }
}


// ----------------------------------------------------
// 3. RDS (PostgreSQL) 用 SG
// ----------------------------------------------------
resource "aws_security_group" "rds" {
  name        = "${var.project_name}-sg-rds"
  vpc_id      =aws_vpc.main.id
  description = "Allows access to RDS from ECS and SSM Bastion."

  // インバウンド１：ECS FargateのSGからPostgreSQLポート(5432)へのアクセスを許可
  ingress {
    description     = "From ECS Fargate"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_fargate.id]
  }

  // インバウンド２：SSM踏み台EC2のSGからPostgreSQLポート(5432)へのアクセスを許可
  // NOTE: SSM踏み台用SGは別途定義しますが、ここではリソースIDで参照することを想定
  // security_groups = [aws_security_group.ssm_bastion.id]
}


// ----------------------------------------------------
// 4. ElastiCache (Redis) 用 SG
// ----------------------------------------------------
resource "aws_security_group" "redis" {
  name        = "${var.project_name}-sg-redis"
  vpc_id      =aws_vpc.main.id
  description = "Allows access to Redis from ECS Fargate."

  // インバウンド：ECS FargateのSGからRedisポート(6379)へのアクセスを許可
  ingress {
    description     = "From ECS Fargate"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_fargate.id]
  }
}

// ----------------------------------------------------
// 5. SSM踏み台サーバー (Bastion) 用 SG
// ----------------------------------------------------
resource "aws_security_group" "ssm_bastion" {
  name        = "${var.project_name}-sg-ssm-bastion"
  vpc_id      = aws_vpc.main.id
  description = "Security group for SSM Bastion Host"

  // インバウンド：なし (SSM接続なのでポート開放不要！これがセキュア)
  
  // アウトバウンド：SSMエージェントがAWSと通信するためにHTTPS(443)が必要
  // また、yum update等やRDSへの接続のため、全許可にしておくのが一般的
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}