# kota-todo

Go 言語で開発した、業務向けリアルタイム通知・タスク管理 Web API のポートフォリオです。

---

## 概要（サマリー）

本プロジェクトは、Go言語とAWSを用いて構築した  
**業務向けリアルタイム通知・タスク管理 Web API** です。

- 想定：中小〜中規模組織の社内業務システム
- 特徴：非同期処理（SQS）と WebSocket によるリアルタイム通知
- 目的：業務委託（週2・フルリモート）における実務対応力の証明

要件定義から設計・実装・テスト・CI/CD・IaC までを一貫して行い、  
**「実運用されること」を前提とした構成**を意識しています。

---

## 技術スタック

### バックエンド
- Go
- Gin
- PostgreSQL
- Redis

### 非同期・リアルタイム処理
- Amazon SQS（DLQ含む）
- WebSocket
- Redis Pub/Sub

### インフラ・運用
- AWS（ECS / RDS / SQS / S3 / IAM / SSM）
- Terraform（IaC）
- Docker / Docker Compose
- GitHub Actions（CI/CD）

---

## アーキテクチャ構成

本システムは、API・非同期処理・リアルタイム通知を疎結合に分離した構成です。

- API サーバー：Gin による REST API
- 非同期処理：SQS + Worker によるバックグラウンド処理
- リアルタイム通知：WebSocket + Redis Pub/Sub
- インフラ管理：Terraform による IaC

詳細な構成図・説明は以下資料に記載しています。

---

## アーキテクチャ・設計方針

- レイヤードアーキテクチャ（Handler / Service / Repository）
- 責務を明確に分離し、テスト容易性を重視
- 非同期処理による API 応答時間の最適化
- 障害を前提とした設計（DLQ / リトライ考慮）

---

## エラーハンドリング方針

- HTTP ステータスコードを厳密に使い分け
- 業務エラーとシステムエラーの明確な分離
- ログ出力を前提としたエラー設計
- 非同期処理失敗時は DLQ に退避し、再処理可能とする

---

## テスト方針

- 認可・権限制御を最重要テスト対象として位置付け
- Service / Repository レイヤー中心のユニットテスト
- 業務ロジックの正常系・異常系を明確に分離

---

## 今後の改善予定

- OpenAPI（Swagger）による API ドキュメント自動生成
    
- 認証機構（JWT / OAuth2）の追加
    
- ページネーション・検索条件の拡張
    
- 監視・アラート設計の強化
    

---

## 自己評価

- Go を用いた業務向け API 設計・実装を一通り経験
    
- 非同期処理・リアルタイム通信を含む構成の設計力
    
- 「動くコード」だけでなく、「運用されるコード」を意識
    
- 週2・リモート業務委託での即戦力を想

---

## ポートフォリオ詳細資料

設計背景・構成図・技術選定理由・トレードオフについては  
以下の資料に詳しくまとめています。

- [https://github.com/Kota-2021/kota-todo/blob/main/docs/portfolio-README.md](https://github.com/Kota-2021/kota-todo/blob/main/docs/portfolio-README.md?utm_source=chatgpt.com)