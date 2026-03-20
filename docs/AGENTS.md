# docs/AGENTS.md

AI向け `docs/` ナビゲーションガイド。ここを起点に必要なドキュメントを特定して参照すること。

---

## プロダクト概要

**KCL Progress Board** — ハッカソン向けSlack Bot + Webダッシュボード。

| 対象ユーザー | ハッカソン参加者（初心者）/ メンター / 運営 |
| --- | --- |
| コア機能 | 進捗可視化（`/progress`）、Q&A AI一次回答（`/question`）、SlackからGitHub Issue自動起票 |
| MVP優先度 | P0のみまず実装。P1以降は体験向上フェーズ |

---

## ファイルマップ

| ファイル | 内容の要点 | 参照すべき場面 |
| --- | --- | --- |
| [`spesfication.md`](./spesfication.md) | Why / What / How の全体概要。課題・スコープ・アイデア・開発ステップ | 仕様の背景・意図を確認したいとき |
| [`details/feature-list.md`](./details/feature-list.md) | 機能一覧と優先度（P0〜P3）。カテゴリA（進捗）・B（Q&A）・C（Issue化）・D（分析） | 機能追加・スコープ判断・優先度確認 |
| [`details/tech-stack.md`](./details/tech-stack.md) | 技術選定一覧（フロント: Next.js/TS/ShadCN、バック: Go/oapi-codegen/sqlc、DB: PostgreSQL、インフラ: AWS ECS/CDK/SQS）| 技術選定理由・依存追加の判断 |
| [`details/screen-flow.md`](./details/screen-flow.md) | Web画面遷移（S-01〜S-05）とSlack操作遷移（L-01〜L-05）。Mermaidフローチャートあり | 画面・UX設計・遷移実装 |
| [`details/permission-design.md`](./details/permission-design.md) | RBAC設計。MVPでは `MENTOR` ロールのみ。ABAC・チームスコープは将来対応 | 認可実装・ミドルウェア設計 |
| [`details/directory.md`](./details/directory.md) | Monorepoディレクトリ構成（`backend/` Go、`web/` Next.js、`lambda/`、`infrastructure/cdk/`）| ファイル配置・新規ファイル作成場所の確認 |
| [`details/infrastructure.md`](./details/infrastructure.md) | AWSインフラ構成（ECS、CloudFront/WAF、SQS、RDB）。Mermaidアーキテクチャ図あり | インフラ変更・デプロイ設計 |
| [`details/process-flow.md`](./details/process-flow.md) | Slack Webhookの処理フロー設計。同期受信（署名検証→SQS投入→200 OK）と非同期Worker処理（progress/question/issueの3キュー）のMermaid Sequence図あり | Slack連携実装・非同期処理設計 |

---

## アーキテクチャ要点（素早い把握用）

```
Slack (/progress, /question, action)
  └→ POST /webhook/slack [署名検証 → 重複チェック → SQS投入 → 200 OK]
       └→ SQS キュー（progress / question / issue + 各DLQ）
            └→ Worker（非同期処理: DB保存 / LLM呼び出し / GitHub API）
                                                                    ↘
Web Dashboard (Next.js) ←→ Go REST API ←→ PostgreSQL
```

- Slackの3秒制約対応のため、受信は即時200 OKを返しSQSで非同期化
- LLM: Amazon Bedrock（Vercel AI SDK + `@ai-sdk/amazon-bedrock` アダプター）
- 認証: Webは独自Session認証（Mentorのみ）、Slack側はワークスペース認証を前提

