# ディレクトリ構成図

---

# 0. 設計前提

| 項目 | 内容 |
| --- | --- |
| リポジトリ構成 | Monorepo |
| アーキテクチャ | Layered / Clean Architecture |
| DDD方針 | 現時点では採用しない（MVPスコープでは過剰な複雑性になるため） |
| デプロイ単位 | 単一サービス |
| 言語 | TypeScript / Go |
| MVP方針 | P0に必要なディレクトリのみ |

---

# 1. 全体構成（Monorepo）

```id="o8d9si"
root/
├── backend/                   # Go backend application
├── lamda/                     # AWS Lambda functions
├── web/                       # Frontend application
├── infrastructure/            # Infrastructure as Code
│   └── terraform/
│       └── aws/
│           ├── modules/
│           └── projects/
├── scripts/                   # Helper scripts
├── docs/
└── README.md
```

---

# 2. フロントエンド構成テンプレ

```id="tq93md"
web/
├── src/
│   ├── app/                   # Routing layer
│   ├── features/              # Feature modules
│   ├── components/            # Shared UI components
│   ├── hooks/
│   ├── lib/                   # API clients
│   ├── stores/                # State management
│   └── types/
├── public/
└── tests/
```

---

# 3. バックエンド構成テンプレ（Go / Clean Architecture）

```id="9wh13c"
backend/
├── api/
│   └── rest/                # HTTP handlers / routers
├── application/             # Use cases / application services
├── entity/                  # Shared domain models
├── cmd/
│   └── serve/
│       └── command.go       # Command definitions (package cmd)
├── main.go                  # Entry point (imports cmd package)
├── database/                # DB schema / migration / query layer
├── infrastucture/           # External integrations and adapters
├── pkg/                     # Shared packages
├── testutil/
│   └── util.go
└── go.mod
```

---

# 4. Lamda構成テンプレ（TypeScript）

```id="kl2m91"
lamda/
├── worker/
│   ├── question.ts
│   ├── question.test.ts
│   ├── issue.ts
│   └── issue.test.ts
├── testutil/
│   └── util.ts
├── package.json
└── tsconfig.json
```

---

# 5. インフラ構成

```id="v9k0mz"
infrastructure/
└── terraform/
    └── aws/
        ├── modules/
        └── projects/
            ├── dev/
            ├── staging/
            └── prod/
```

---

# 6. テスト構成テンプレ

```id="c32po1"
backend/
├── application/
│   └── ...
│       ├── service.go
│       └── service_test.go
├── entity/
│   └── ...
│       ├── model.go
│       └── model_test.go
├── api/
│   └── rest/
│       └── ...
│           ├── handler.go
│           └── handler_test.go
└── testutil/
    └── util.go

lamda/
├── worker/
│   └── ...
│       ├── question.ts
│       ├── question.test.ts
│       ├── issue.ts
│       └── issue.test.ts
└── testutil/
    └── util.ts
```

- Integration/E2E専用ディレクトリは作成しない
- backendテストは対象実装ファイルと同階層に `*_test.go` を配置する
- lamdaテストは対象実装ファイルと同階層に `*.test.ts` を配置する
