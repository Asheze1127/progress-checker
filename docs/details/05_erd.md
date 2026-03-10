# 🗄️ DB設計書：KCL Progress Board

---

# 0️⃣ 設計観点

| 項目      | 内容                                    |
| ------- | ------------------------------------- |
| 権限モデル   | RBAC（単一ロール: MENTOR）                  |
| ID戦略    | UUID（全テーブル共通）                         |
| 論理削除    | 無（MVP）                                |
| 監査ログ    | 任意（MVP）                               |
| ORM     | sqlc（型安全なDBアクセス）                      |

---

# 1️⃣ テーブル一覧

| ドメイン  | テーブル名               | 役割                               | Phase |
| ----- | ------------------- | -------------------------------- | ----- |
| アカウント | users               | ユーザー主体（参加者・メンター）                | P0    |
| 組織    | teams               | ハッカソンチーム                        | P0    |
| 組織    | team_members        | ユーザーとチームの所属関係                   | P0    |
| 組織    | mentor_assignments  | メンターの担当チーム紐付け                   | P2    |
| コア機能  | progress_logs       | `/progress` による進捗投稿ログ           | P0    |
| コア機能  | questions           | `/question` による質問                | P0    |
| コア機能  | question_sessions   | AI会話セッション管理（followup判定）         | P1    |
| 認証    | tokens              | Web認証トークン                        | P0    |
| 補助    | idempotency_keys    | Slackイベント重複排除（Goインメモリキャッシュで管理） | —     |

---

# 2️⃣ ERD

```mermaid
erDiagram
    users {
        uuid id PK
        varchar slack_user_id UK
        varchar name
        varchar email
        varchar role "participant | mentor"
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    teams {
        uuid id PK
        varchar name
        varchar slack_channel_id
        timestamp created_at
        timestamp updated_at
    }

    team_members {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        timestamp joined_at
    }

    mentor_assignments {
        uuid id PK
        uuid mentor_id FK
        uuid team_id FK
        timestamp created_at
    }

    progress_logs {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        varchar phase "idea | design | coding | testing | demo"
        boolean sos
        text comment
        varchar slack_msg_ts
        timestamp created_at
    }

    questions {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        varchar title
        text body
        varchar status "open | in_progress | resolved"
        varchar slack_thread_ts
        varchar slack_channel_id
        uuid assigned_to FK "nullable"
        varchar github_issue_url "nullable"
        timestamp created_at
        timestamp updated_at
    }

    question_sessions {
        uuid id PK
        uuid question_id FK
        varchar status "awaiting_ai | awaiting_user | escalated | resolved"
        int max_follow_ups
        int current_round
        timestamp created_at
        timestamp updated_at
    }

    tokens {
        uuid id PK
        uuid user_id FK
        varchar token
        timestamp expires_at
        timestamp created_at
    }


    users ||--o{ team_members : "belongs to"
    teams ||--o{ team_members : "has"
    users ||--o{ mentor_assignments : "mentors"
    teams ||--o{ mentor_assignments : "mentored by"
    teams ||--o{ progress_logs : "logs"
    users ||--o{ progress_logs : "submits"
    teams ||--o{ questions : "asks"
    users ||--o{ questions : "asks"
    users ||--o{ questions : "assigned_to"
    questions ||--|| question_sessions : "has session"
    users ||--o{ tokens : "authenticates"
```

---

# 3️⃣ カラム定義

## users

| カラム          | 型         | 制約              | 説明                          |
| ------------ | --------- | --------------- | --------------------------- |
| id           | UUID      | PK              |                             |
| slack_user_id| VARCHAR   | UNIQUE NOT NULL | Slack上のユーザーID               |
| name         | VARCHAR   | NOT NULL        | 表示名                         |
| email        | VARCHAR   |                 | メールアドレス（任意）                 |
| role         | VARCHAR   | NOT NULL        | participant / mentor        |
| is_active    | BOOLEAN   | NOT NULL DEFAULT true | アカウント有効判定           |
| created_at   | TIMESTAMP | NOT NULL        |                             |
| updated_at   | TIMESTAMP | NOT NULL        |                             |

---

## teams

| カラム             | 型         | 制約              | 説明                   |
| --------------- | --------- | --------------- | -------------------- |
| id              | UUID      | PK              |                      |
| name            | VARCHAR   | NOT NULL        | チーム名                 |
| slack_channel_id| VARCHAR   |                 | チーム用Slackチャンネル       |
| created_at      | TIMESTAMP | NOT NULL        |                      |
| updated_at      | TIMESTAMP | NOT NULL        |                      |

---

## team_members

| カラム      | 型         | 制約       | 説明         |
| -------- | --------- | -------- | ---------- |
| id       | UUID      | PK       |            |
| team_id  | UUID      | FK(teams)| 所属チーム      |
| user_id  | UUID      | FK(users)| 所属ユーザー     |
| joined_at| TIMESTAMP | NOT NULL | 参加日時       |

UNIQUE(team_id, user_id)

---

## mentor_assignments

| カラム       | 型         | 制約       | 説明           |
| --------- | --------- | -------- | ------------ |
| id        | UUID      | PK       |              |
| mentor_id | UUID      | FK(users)| メンター         |
| team_id   | UUID      | FK(teams)| 担当チーム        |
| created_at| TIMESTAMP | NOT NULL |              |

UNIQUE(mentor_id, team_id)

---

## progress_logs

| カラム         | 型         | 制約        | 説明                                   |
| ----------- | --------- | --------- | ------------------------------------ |
| id          | UUID      | PK        |                                      |
| team_id     | UUID      | FK(teams) | 投稿チーム                                |
| user_id     | UUID      | FK(users) | 投稿者                                  |
| phase       | VARCHAR   | NOT NULL  | idea / design / coding / testing / demo |
| sos         | BOOLEAN   | NOT NULL DEFAULT false | SOSフラグ                    |
| comment     | TEXT      |           | コメント（自由記述）                           |
| slack_msg_ts| VARCHAR   |           | Slack投稿のタイムスタンプ                      |
| created_at  | TIMESTAMP | NOT NULL  |                                      |

---

## questions

| カラム             | 型         | 制約        | 説明                            |
| --------------- | --------- | --------- | ----------------------------- |
| id              | UUID      | PK        |                               |
| team_id         | UUID      | FK(teams) | 質問元チーム                        |
| user_id         | UUID      | FK(users) | 質問者                           |
| title           | VARCHAR   | NOT NULL  | 質問タイトル                        |
| body            | TEXT      | NOT NULL  | 質問本文（テンプレート形式）                |
| status          | VARCHAR   | NOT NULL  | open / in_progress / resolved |
| slack_thread_ts | VARCHAR   |           | Slackスレッドのタイムスタンプ             |
| slack_channel_id| VARCHAR   |           | 投稿先チャンネルID                    |
| assigned_to     | UUID      | FK(users) NULLABLE | 担当メンター               |
| github_issue_url| VARCHAR   | NULLABLE  | Issue化した場合のURL                |
| created_at      | TIMESTAMP | NOT NULL  |                               |
| updated_at      | TIMESTAMP | NOT NULL  |                               |

---

## question_sessions

| カラム           | 型         | 制約             | 説明                                                |
| ------------- | --------- | -------------- | ------------------------------------------------- |
| id            | UUID      | PK             |                                                   |
| question_id   | UUID      | FK(questions) UNIQUE | 1質問に1セッション                                  |
| status        | VARCHAR   | NOT NULL       | awaiting_ai / awaiting_user / escalated / resolved |
| max_follow_ups| INT       | NOT NULL DEFAULT 3 | AI最大フォローアップ回数                             |
| current_round | INT       | NOT NULL DEFAULT 0 | 現在のラウンド数                                    |
| created_at    | TIMESTAMP | NOT NULL       |                                                   |
| updated_at    | TIMESTAMP | NOT NULL       |                                                   |

---

## tokens

| カラム       | 型         | 制約        | 説明           |
| --------- | --------- | --------- | ------------ |
| id        | UUID      | PK        |              |
| user_id   | UUID      | FK(users) | トークン所有者      |
| token     | VARCHAR   | NOT NULL  | JWTトークン文字列   |
| expires_at| TIMESTAMP | NOT NULL  | 有効期限         |
| created_at| TIMESTAMP | NOT NULL  |              |

---


# 4️⃣ 権限設計

## RBAC（MVP）

- `users.role` で判定
- Web画面へのアクセスは `role = mentor` のみ許可
- Slack操作は全ユーザーが利用可能

## 将来拡張（ABAC）

- `mentor_assignments` テーブルで担当チームスコープ制御を追加予定

```pseudo
if user.role != "mentor":
    deny (403)
if resource.team_id not in user.assigned_team_ids:
    deny (403)
allow
```

---

# 5️⃣ インデックス設計

| テーブル             | カラム                      | 種別     | 用途                         |
| ---------------- | ------------------------ | ------ | -------------------------- |
| users            | slack_user_id            | UNIQUE | Slack→ユーザー検索               |
| team_members     | (team_id, user_id)       | UNIQUE | 重複所属防止                     |
| mentor_assignments| (mentor_id, team_id)    | UNIQUE | 重複アサイン防止                   |
| progress_logs    | (team_id, created_at)    | INDEX  | チーム別進捗の時系列取得               |
| questions        | status                   | INDEX  | ステータス別キュー取得                |
| questions        | (team_id, created_at)    | INDEX  | チーム別質問の時系列取得               |
| question_sessions| question_id              | UNIQUE | 1質問1セッション制約                |
| tokens           | token                    | INDEX  | トークン検証                     |

---

# 6️⃣ 設計上の意思決定ログ

## ADR-001: Slackイベント重複排除をDBではなくGoインメモリキャッシュで管理する

| 項目 | 内容 |
| --- | --- |
| **決定日** | 2026-03-05 |
| **ステータス** | 採用 |

### 背景

Slackの [Events API](https://docs.slack.dev/apis/events-api/#responding) はイベントを重複送信することがあるため、冪等性の担保が必要。
当初は `idempotency_keys` テーブルをDBで管理する案を検討した。

### 検討した選択肢

| 案 | 概要 |
| --- | --- |
| RDB（`idempotency_keys` テーブル） | 永続化できるが、WebhookのたびにDB書き込みが発生する。定期パージも必要 |
| Goインメモリキャッシュ（`sync.Map` + TTL） | 追加インフラ不要。軽量で3秒制約に有利 |
| Redis（ElastiCache） | TTL管理が容易だが、インフラコストと複雑性が増す |

### 決定内容

**Goインメモリキャッシュ（`sync.Map` + TTL）を採用する。**

### 理由

- ECS上でコンテナが長期起動するため、再起動による消失リスクは実質低い
- Slackの重複イベントは数秒〜数分以内に届くケースがほとんどであり、TTLをそれに合わせれば十分
- スループットはピーク数件/分程度（仕様書より）のため、DB書き込みのボトルネックは起きないが、シンプルさを優先する
- Redis追加はインフラコストと管理コストが見合わない（ハッカソン規模）

### トレードオフ・注意点

- コンテナが複数台にスケールした場合、インスタンス間でキャッシュが共有されないため重複を取りこぼす可能性がある（現時点では単一コンテナ運用のため許容）
