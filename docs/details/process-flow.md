# 🔄 処理フロー設計（MVP / SQSパターン）

---

## 0. 設計前提

| 項目 | 内容 |
| --- | --- |
| 対象機能 | `/progress`、`/question`、Slack議論のIssue化、Webダッシュボード表示 |
| イベント起点 | Slack Slash Command / Slack Action |
| 受信方式 | `POST /webhook/slack` で全イベントを受信 |
| 非同期方針 | 受信処理は署名検証・重複チェック・SQS投入までで終了し、3秒以内に `200 OK` を返却 |
| 実行方式 | SQS Consumer（Worker）が非同期で本処理を実行 |
| キュー構成 | `progress` / `question` / `issue` の3キュー（各キューにDLQ設定） |
| question分岐 | `thread_ts` + `question_sessions` を使って「新規質問 / 継続回答」を判定 |
| 主要コンポーネント | Slack、Webhook API、Idempotency Store、Question Session Store、SQS、Worker、PostgreSQL、LLM、GitHub API |

---

## 1. フェーズ分割

### 1.1 受信フェーズ（同期・ここでHTTP処理は終了）

```mermaid
flowchart LR
    S[Slack] --> API[Webhook API /webhook/slack]
    API --> V{Slack署名検証}
    V -->|NG| E1[401/403 返却]
    V -->|OK| D{重複イベント?}
    D -->|Yes| A1[200 OK 返却]
    D -->|No| R{イベント種別}

    R -->|progress| QP[SQS progress へ送信]

    R -->|question| QCHK{questionは継続回答か?}
    QCHK -->|No| QN[SQS question:new へ送信]
    QCHK -->|Yes| QF[SQS question:followup へ送信]

    R -->|issue| QI[SQS issue へ送信]

    QP --> A2[200 OK 返却]
    QN --> A2
    QF --> A2
    QI --> A2
```

### 1.2 実行フェーズ（非同期・受信後に別処理として実行）

```mermaid
flowchart LR
    QP[SQS progress] --> PW[Progress Worker]
    QQ[SQS question<br/>type=new/followup] --> QW[Question Worker]
    QI[SQS issue] --> IW[Issue Worker]

    PW --> DS1[処理完了]
    QW --> DS2[処理完了]
    IW --> DS3[処理完了]

    PW -.失敗.-> RP[SQS再試行]
    QW -.失敗.-> RQ[SQS再試行]
    IW -.失敗.-> RI[SQS再試行]

    RP --> DLP[progress DLQ]
    RQ --> DLQ[question DLQ]
    RI --> DLI[issue DLQ]
```

---

## 2. `/progress` フロー

### 2.1 受信フロー（同期）

```mermaid
sequenceDiagram
    participant User as Participant
    participant Slack as Slack
    participant API as Webhook API
    participant IDS as Idempotency Store
    participant SQSP as SQS progress

    User->>Slack: /progress 入力
    Slack->>API: command payload
    API->>API: 署名検証 + 必須項目チェック
    API->>IDS: idempotency_key 登録確認
    API->>SQSP: progress message 送信
    API-->>Slack: 200 OK（3秒以内）
```

### 2.2 実行フロー（非同期）

```mermaid
sequenceDiagram
    participant SQSP as SQS progress
    participant W as Progress Worker
    participant DB as PostgreSQL
    participant Slack as Slack
    participant Dash as Dashboard API

    SQSP->>W: message delivery
    W->>DB: 進捗データ保存
    W->>Slack: #progress-board に整形投稿
    W->>Dash: ダッシュボード反映
    W-->>SQSP: ack（delete message）
```

---

## 3. `/question` フロー

### 3.1 受信フロー（同期）

```mermaid
sequenceDiagram
    participant User as Participant
    participant Slack as Slack
    participant API as Webhook API
    participant IDS as Idempotency Store
    participant QSS as Question Session Store
    participant SQSQ as SQS question

    User->>Slack: /question 入力 or スレッド返信
    Slack->>API: command/event payload
    API->>API: 署名検証
    API->>IDS: idempotency_key 登録確認

    alt 新規質問
        API->>SQSQ: type=question_new message 送信
    else 継続回答（thread_tsあり + awaiting_user）
        API->>QSS: session状態を確認
        API->>SQSQ: type=question_followup message 送信
    end

    API-->>Slack: 200 OK（3秒以内）
```

### 3.2 実行フロー（非同期）

```mermaid
sequenceDiagram
    participant SQSQ as SQS question
    participant W as Question Worker
    participant QSS as Question Session Store
    participant Slack as Slack
    participant LLM as LLM API
    participant Mentor as Mentor Channel

    SQSQ->>W: message delivery

    alt type=question_new
        W->>LLM: 初回質問を送信
    else type=question_followup
        W->>Slack: thread_ts の履歴を取得
        W->>QSS: 過去の不足情報リクエストを取得
        W->>LLM: 履歴 + 追加回答をまとめて送信
    end

    LLM-->>W: answer / follow-up
    W->>Slack: スレッド返信

    alt follow-up要求が必要
        W->>QSS: status=awaiting_user に更新
    else 解決
        W->>QSS: status=resolved に更新
    end

    alt 解決しない and メンター対応が必要
        W->>Mentor: メンション付き転送
    end

    W-->>SQSQ: ack（delete message）
```

---

## 4. Slack議論のIssue化フロー

### 4.1 受信フロー（同期）

```mermaid
sequenceDiagram
    participant User as Participant
    participant Slack as Slack
    participant API as Webhook API
    participant IDS as Idempotency Store
    participant SQSI as SQS issue

    User->>Slack: Issue化アクション
    Slack->>API: action payload
    API->>API: 署名検証
    API->>IDS: idempotency_key 登録確認
    API->>SQSI: issue message 送信
    API-->>Slack: 200 OK
```

### 4.2 実行フロー（非同期）

```mermaid
sequenceDiagram
    participant SQSI as SQS issue
    participant W as Issue Worker
    participant Slack as Slack
    participant LLM as LLM API
    participant GH as GitHub API

    SQSI->>W: message delivery
    W->>Slack: スレッド履歴取得
    W->>LLM: 要約（背景/課題/Done条件）生成
    W->>GH: Issue作成
    GH-->>W: Issue URL
    W->>Slack: Issue作成完了を通知
    W-->>SQSI: ack（delete message）
```

---

## 5. エラーフロー分離

### 5.1 受信フェーズの失敗（同期）

```mermaid
flowchart LR
    A[Webhook受信] --> B{署名検証}
    B -->|失敗| C[401/403 即時返却]
    B -->|成功| D{SQS送信成功?}
    D -->|Yes| E[200 OK]
    D -->|No| F[5xx返却]
```

### 5.2 実行フェーズの失敗（非同期）

```mermaid
flowchart LR
    A[SQS message delivery] --> B[Worker処理]
    B --> C{処理成功?}
    C -->|Yes| D[message delete]
    C -->|No| E[visibility timeout後に再配信]
    E --> F{maxReceiveCount超過?}
    F -->|No| A
    F -->|Yes| G[DLQへ移送 + 運営通知]
```

