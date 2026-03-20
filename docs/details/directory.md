# ディレクトリ構成図

---

# 0. 設計前提

| 項目           | 内容                                                          |
| -------------- | ------------------------------------------------------------- |
| リポジトリ構成 | Monorepo                                                      |
| アーキテクチャ | Layered / Clean Architecture                                  |
| DDD方針        | 現時点では採用しない（MVPスコープでは過剰な複雑性になるため） |
| デプロイ単位   | 単一サービス                                                  |
| 言語           | TypeScript / Go                                               |
| MVP方針        | P0に必要なディレクトリのみ                                    |

---

# 1. 全体構成（Monorepo）

```id="o8d9si"
root/
├── backend/                   # Go backend application
├── lambda/                    # AWS Lambda functions
├── web/                       # Frontend application
├── infrastructure/            # Infrastructure as Code
│   └── cdk/
│       ├── bin/
│       ├── lib/
│       │   ├── stacks/
│       │   └── constructs/
│       └── test/
├── scripts/                   # Helper scripts
├── docs/
└── README.md
```

---

# 2. フロントエンド構成テンプレ

```id="tq93md"
web/
├── .storybook/                    # Storybook configuration
├── public/                        # Static files
├── src/
│   ├── app/                       # Next.js App Router
│   │   ├── page.tsx               # Root page
│   │   ├── layout.tsx             # Common layout
│   │   ├── globals.css            # Global styles
│   │   ├── (group)/               # Route Group
│   │   │   └── xxx/
│   │   │       └── page.tsx
│   │   └── api/                   # BFF entry points
│   │       └── xxx/
│   │           └── route.ts
│   ├── features/                  # Feature modules
│   │   └── xxx/
│   │       ├── container/         # Screen logic
│   │       │   └── XxxContainer.tsx
│   │       ├── element/           # Presentational components
│   │       │   └── XxxElement.tsx
│   │       ├── model/             # Data definitions for the feature
│   │       │   └── xxx.ts
│   │       ├── utils/             # Helper functions for the feature
│   │       │   └── xxx.ts
│   │       └── index.ts           # Public API of the feature
│   ├── components/                # Shared UI components
│   │   ├── ui/                    # shadcn/ui and common UI components
│   │   ├── layout/                # Common layout components
│   │   └── feedback/              # Error, empty state, and success displays
│   ├── hooks/                     # Generic hooks only
│   ├── lib/                       # API clients
│   │   ├── fetcher/               # Shared HTTP client for all features
│   │   ├── swr/                   # Common SWR configuration
│   │   ├── constants/             # Constants
│   │   ├── utils/                 # Shared utility functions
│   │   ├── auth/                  # Shared auth logic
│   │   ├── i18n/                  # Placeholder for future i18n support
│   │   └── providers/             # React providers
│   ├── server/                    # BFF internal implementation
│   │   ├── bff/                   # Per-API service logic
│   │   ├── client/                # External API / backend clients
│   │   ├── auth/                  # Server-side auth logic
│   │   └── utils/                 # Server-side utility functions
│   ├── styles/                    # Style-related files
│   ├── test/                      # Test configuration and mocks
│   ├── types/                     # App-wide type definitions
│   └── stories/                   # Storybook documentation
├── components.json                # shadcn/ui configuration
├── package.json
├── tsconfig.json
├── next.config.js
└── README.md
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
├── infrastructure/          # External integrations and adapters
├── pkg/                     # Shared packages
├── testutil/
│   └── util.go
└── go.mod
```

---

# 4. Lambda構成テンプレ（TypeScript）

```id="kl2m91"
lambda/
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
└── cdk/
    ├── bin/
    │   └── app.ts             # CDK app entry point
    ├── lib/
    │   ├── stacks/            # Stack definitions per environment
    │   │   ├── dev.ts
    │   │   ├── staging.ts
    │   │   └── prod.ts
    │   └── constructs/        # Reusable CDK constructs
    ├── test/                  # CDK unit tests
    ├── cdk.json
    ├── package.json
    └── tsconfig.json
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

lambda/
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
- lambdaテストは対象実装ファイルと同階層に `*.test.ts` を配置する
