# AGENTS.md — KCL Progress Board

このリポジトリで coding agent（Codex / Copilot Coding Agent / Cursor など）が作業するための指示書。
README は人間向け、本ファイルは「エージェントが迷いやすい点」のみを記載する。

---

## 0) ゴール（このプロダクトが解決する課題と提供する機能）

### 解決する課題

- 参加者同士で進捗が見えず、切磋琢磨しづらい
- 質問の出し方・場所が分からず、質問が出ない or メンターに分散して埋もれる
- Slack 上の有益な議論がフロー情報として流れてしまい、タスクとして残らない

### 提供する 3 機能

1. **進捗管理**: Slack `/progress` で進捗投稿 → DB 保存 → Web ダッシュボードに一覧表示
2. **Q&A 自動化**: Slack `/question` で質問 → LLM が一次回答＋不足情報の聞き返し → 必要ならメンター ch へ転送
3. **Issue 化**: Slack スレッドを要約 → GitHub Issue として自動起票（手動トリガー：絵文字 or メニュー）

### MVP 開発ステップ（この順番で実装する）

| Step | 内容 | 概要 |
|------|------|------|
| 1 | 進捗機能 | `/progress` → フォーマット入力 → Webhook → DB保存 → Web表示 |
| 2 | Q&A基盤 | `/question` → Bot投稿 → LLM一次回答＋ヒアリング |
| 3 | Issue化 | Slackアクション検知 → スレッド取得 → LLM要約 → GitHub Issue起票 |

---

## 1) 技術スタック・起動コマンド

### 技術スタック

| レイヤー | 技術 | 補足 |
|----------|------|------|
| Backend | Go | API サーバー。Slack Event/Slash Command 受け口、DB操作、LLM呼び出し、GitHub API |
| Frontend | React | Web ダッシュボード UI（シンプルな一覧画面） |
| DB | PostgreSQL | チーム情報・進捗ログなどを構造化管理。AI会話ログなど柔軟性が必要な部分は JSONB で対応 |
| Infra | Cloud Run（想定） | コンテナベース |
| LLM | 未定 | 一次回答・要約・Issue化に使用 |

### 起動コマンド

```bash
# Backend
make backend-dev      # ローカル起動
go test ./...         # テスト
golangci-lint run     # リンタ

# Frontend
make frontend-dev     # ローカル起動
npm test              # テスト
npm run lint          # リンタ

# DB
docker compose up -d db

# 全体
make dev              # 全サービス起動
make test             # 全テスト実行
```

---

## 2) リポジトリ構造

> エージェントが最も迷う箇所。実態と異なる場合はここを最優先で修正すること。

```
root/
├── backend/           # Go API（Slack Events / Slash Command / DB / LLM / GitHub API）
│   ├── cmd/           # エントリポイント
│   ├── internal/      # Clean Architecture（domain / usecase / repository / handler / middleware / config）
│   ├── migrations/    # DBマイグレーション
│   └── tests/
├── frontend/          # React Web ダッシュボード
│   ├── src/
│   │   ├── app/       # ルーティング
│   │   ├── features/  # 機能単位（progress / question / issue）
│   │   ├── components/# 共通 UI
│   │   ├── hooks/
│   │   ├── lib/       # API クライアント
│   │   └── types/
│   └── tests/
├── docs/              # 仕様・運用メモ
│   ├── spesfication.md    # プロダクト仕様
│   └── details/           # 詳細設計テンプレート群
├── scripts/           # 開発補助（マイグレーション、ローカル起動など）
├── .github/           # Issue/PR テンプレート、CI設定
├── AGENTS.md          # 本ファイル（エージェント向け指示）
└── README.md          # 人間向け概要
```

---

## 3) エージェントのワークフロー

ユーザーからの指示キーワードに応じて作業モードを切り替える。

| 指示 | やること | やらないこと |
|------|----------|------------|
| 「調査」 | 調査結果を `docs/reports/` に記載 | コードを書く |
| 「計画」 | `tasks.md` に計画を記載。コードベースと docs を読み、関連ファイルパスをすべて列挙。不明点は fetch MCP で検索。必要最小限の要件のみ記載 | コードを書く |
| 「実装」 | `tasks.md` の内容に基づいて実装 | 記載以上の実装、デバッグ |
| 「デバッグ」 | 直前のタスクのデバッグ「手順」のみ提示 | 勝手にコードを修正 |

### 共通ルール

- 日本語で応答する
- 絵文字を多めに使用する
- 必要に応じてユーザーに質問し、要求を明確にする
- 作業後、作業内容とユーザーが次に取れる行動を説明する
- 作業項目が多い場合は段階に区切り、git commit しながら進める（semantic commit を使用）
- コマンドの出力が確認できない場合、get last command / check background terminal で確認する

---

## 4) ルール（3 段階の境界）

### ✅ Always do（必ずやる）

- 変更は小さく：**1PR = 1目的**。不要な整形はしない
- 変更した領域のテスト・リンタを必ず実行する（Backend → `go test` / Frontend → `npm test`）
- Slack の入出力フォーマット（`/progress` `/question`）は既存テンプレを壊さない（互換性優先）
- 秘密情報は **絶対にコミットしない**（`.env` / token / signing secret / webhook URL など）
- エラーハンドリング：Slack へ返すメッセージには「次に何を出せばよいか（不足情報）」を必ず含める
- Go のエラーは呼び出し元に返す。握りつぶさない

### ⚠️ Ask first（事前確認して）

- 依存追加（npm / Go module）や大きなライブラリ導入
- DB スキーマ変更・マイグレーション追加
- Slack アプリの権限（scopes）変更、OAuth フロー変更
- GitHub Issue の起票先リポジトリやラベル体系の変更
- LLM のモデル・プロンプトの大幅変更（誤回答リスクに直結）

### 🚫 Never do（やらない）

- 本番デプロイ・本番設定変更（このリポジトリ内で完結しない操作）
- DB 破壊操作（DROP / TRUNCATE）やデータ消去を勝手に実行
- 「テストを消して通す」「ログを消して見えなくする」方向の修正
- トークンやペイロードの全文ログ出力（マスク必須）

---

## 5) Slack 連携の実装注意

Slack 連携はバグが発生しやすい領域。以下を必ず守る。

| 項目 | ルール |
|------|--------|
| 署名検証 | 全リクエストで `SLACK_SIGNING_SECRET` による署名検証を行う |
| レスポンス | 即時 `200 ack` → 必要なら後続で非同期投稿（3秒タイムアウトに注意） |
| ログ | トークン・ペイロード全文はログ出力禁止（マスクする） |
| 投稿先 | 進捗は `#progress-board` チャンネルに自動投稿 |

### 環境変数

```
SLACK_SIGNING_SECRET    # Slack アプリの署名シークレット
SLACK_BOT_TOKEN         # Bot User OAuth Token（xoxb-）
SLACK_APP_TOKEN         # Socket Mode 用トークン（xapp-）（Socket Mode の場合のみ）
GITHUB_TOKEN            # GitHub API 用トークン（Issue 起票に使用）
DATABASE_URL            # PostgreSQL 接続文字列
LLM_API_KEY             # LLM API キー
PUBLIC_BASE_URL         # 公開 URL（ローカルなら ngrok 等）
```

---

## 6) LLM 連携の実装注意（Q&A / 要約 / Issue 化）

| 項目 | ルール |
|------|--------|
| 一次回答 | 「参考」として提示し、断定しない。不確かな場合は **質問（不足情報）を優先** |
| Issue 化 | 「要約」「決定事項」「次アクション」「未決事項」を分離して記載 |
| 元スレッド | URL / メッセージリンクを必ず残す（追跡可能にする） |
| エラー時 | LLM 呼び出し失敗時、ユーザーに「AIが現在利用できません。メンターに直接質問してください」等のフォールバックメッセージを返す |

---

## 7) Definition of Done（完了条件）

PR をマージしてよい条件：

1. ローカルで該当テスト・リンタが通る
2. 例外系（Slack 署名 NG / 入力不足 / LLM 失敗 / GitHub API 失敗）でユーザーに分かるメッセージが返る
3. `docs/` に運用上の注意が必要なら追記済み（「どこを設定すれば動くか」が分かること）
4. 秘密情報がコミットに含まれていない

---

## 8) プロジェクト情報

| 項目 | 内容 |
|------|------|
| チーム | 2名体制（A: 太郎, B: 次郎） |
| 期間 | 2026年3月 |
| 仕様書 | `docs/spesfication.md` |
| 詳細設計 | `docs/details/` 以下 |
| Issue テンプレート | `.github/ISSUE_TEMPLATE/` |
| PR テンプレート | `.github/pull_request_template.md` |
