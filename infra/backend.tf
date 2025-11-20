terraform {
  // S3バックエンドの設定
  backend "s3" {
    // 1. Stateファイルの保存先バケット名
    bucket = "tfstate-go-realtime-task-api-prod-202511"

    // 2. Stateファイルのパス (本番環境を想定したパス)
    key    = "prod/go-realtime-task-api/terraform.tfstate"

    // 3. リージョン
    region = "ap-northeast-1"

    // 4. State Lock用のDynamoDBテーブル名
    dynamodb_table = "terraform-lock-table"

    // 5. SSL/TLSによる暗号化を有効化
    encrypt = true

    // 6.ローカル環境のプロファイルを明示的に指定
    profile        = "my-portfolio-admin"
  }
}
