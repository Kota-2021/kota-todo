# infra/bastion.tf

# ----------------------------------------------------
# 1. IAM Role for SSM (EC2インスタンスがSSMを利用するためのロール)
# ----------------------------------------------------
resource "aws_iam_role" "ssm_instance" {
  name = "${var.project_name}-ssm-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })

  tags = {
    Name = "${var.project_name}-ssm-role"
  }
}

# ----------------------------------------------------
# 2. IAM Policy Attachment (SSMの管理権限付与)
# ----------------------------------------------------
resource "aws_iam_role_policy_attachment" "ssm_managed_instance" {
  role       = aws_iam_role.ssm_instance.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# ----------------------------------------------------
# 3. IAM Instance Profile (EC2にロールをアタッチするためのコンテナ)
# ----------------------------------------------------
resource "aws_iam_instance_profile" "ssm_instance" {
  name = "${var.project_name}-ssm-profile"
  role = aws_iam_role.ssm_instance.name
}

# ----------------------------------------------------
# 4. EC2 Instance (SSM Bastion Host)
# ----------------------------------------------------

# 最新のAmazon Linux 2 AMIを取得
data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "ssm_bastion" {
  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = "t3.micro" 
  
  # プライベートサブネットに配置
  subnet_id = aws_subnet.private_a.id 
  
  # パブリックIPは不要
  associate_public_ip_address = false 
  
  # SSM用のSGを適用
  security_groups = [aws_security_group.ssm_bastion.id]

  # 作成したIAMインスタンスプロファイルをアタッチ
  iam_instance_profile = aws_iam_instance_profile.ssm_instance.name
  
  tags = {
    Name = "${var.project_name}-ssm-bastion"
  }
  
}