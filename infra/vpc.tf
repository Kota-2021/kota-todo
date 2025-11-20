// プロジェクト名とCIDRを定義（variables.tfからの読み込みを想定）
// 例: var.project_name = "go-realtime-task-api"
// 例: var.vpc_cidr = "10.0.0.0/16"

// ----------------------------------------------------
// 1. VPC
// ----------------------------------------------------
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
  Name = "${var.project_name}-vpc"
}
}

// ----------------------------------------------------
// 2. インターネットゲートウェイ (IGW)
// ----------------------------------------------------
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "${var.project_name}-igw"
  }
}

// ----------------------------------------------------
// 3. アベイラビリティゾーン (AZs)
// ----------------------------------------------------
// 今回は2つのAZを使用
data "aws_availability_zones" "available" {
  state = "available"
}

// ----------------------------------------------------
// 4. サブネットの定義 (Multi-AZ)
// ----------------------------------------------------

// AZ-A のパブリックサブネット
resource "aws_subnet" "public_a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24" // 例
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true // パブリックIP自動割り当て
  tags = {
    Name = "${var.project_name}-public-a"
  }
}

// AZ-B のパブリックサブネット
resource "aws_subnet" "public_b" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.2.0/24" // 例
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = true
  tags = {
    Name = "${var.project_name}-public-b"
  }
}

// AZ-A のプライベートサブネット
resource "aws_subnet" "private_a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.11.0/24" // 例
  availability_zone       = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "${var.project_name}-private-a"
  }
}

// AZ-B のプライベートサブネット
resource "aws_subnet" "private_b" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.12.0/24" // 例
  availability_zone       = data.aws_availability_zones.available.names[1]
  tags = {
    Name = "${var.project_name}-private-b"
  }
}

// ----------------------------------------------------
// 5. NAT Gateway (プライベートからのインターネットアクセス用)
// ----------------------------------------------------
// コスト最適化のため、AZ-Aのパブリックサブネットに1つだけ配置する (シングルNAT構成)

// Elastic IPアドレス
resource "aws_eip" "nat_a" {
  domain     = "vpc"
  depends_on = [aws_internet_gateway.main]
  tags = { Name = "${var.project_name}-nat-a-eip" }
}

// NAT Gateway
resource "aws_nat_gateway" "a" {
  allocation_id = aws_eip.nat_a.id
  subnet_id     = aws_subnet.public_a.id
  tags = {
    Name = "${var.project_name}-nat-a"
  }
}

// ----------------------------------------------------
// 6. ルートテーブル
// ----------------------------------------------------

// パブリックルートテーブル (IGW経由でインターネットへ)
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = { Name = "${var.project_name}-public-rt" }
}

// プライベートルートテーブル (NAT Gateway経由でインターネットへ)
resource "aws_route_table" "private_a" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.a.id
  }
  tags = { Name = "${var.project_name}-private-a-rt" }
}

// ルートテーブルとサブネットの関連付け
resource "aws_route_table_association" "public_a" {
  subnet_id      = aws_subnet.public_a.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private_a" {
  subnet_id      = aws_subnet.private_a.id
  route_table_id = aws_route_table.private_a.id
}

resource "aws_route_table_association" "public_b" {
subnet_id      = aws_subnet.public_b.id
route_table_id = aws_route_table.public.id
}

// (注: ここでは節約のため、AZ-Aにある NAT Gateway を共用します)
resource "aws_route_table_association" "private_b" {
subnet_id      = aws_subnet.private_b.id
route_table_id = aws_route_table.private_a.id
}