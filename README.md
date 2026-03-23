# progress-checker

KCL Progress Board の設計・開発を進めるためのリポジトリです。Slack Bot と Web ダッシュボードを組み合わせて、ハッカソン参加者の進捗可視化、質問の一次対応、Slack 上の議論の Issue 化を支援することを目指しています。

## セットアップ

### 推奨: Dev Container を使う

1. Docker と VS Code Dev Containers 拡張、または GitHub Codespaces を利用できる状態にします。
2. このリポジトリを開き、`Reopen in Container` を実行します。
3. 初回作成時に [`.devcontainer/on-create.sh`](./.devcontainer/on-create.sh) が走り、`.tool-versions` をもとにツールをセットアップします。

4. PostgreSQL コンテナも同時に立ち上がります。

5. AWS を使う作業に入る場合は、コンテナ内で認証を済ませます。

```bash
aws configure
# or
aws sso login --profile <profile>
```

### Dev Container を使わない場合

ローカル環境で進める場合は [`.tool-versions`](./.tool-versions) に合わせて `asdf` などで Node.js / Go / direnv を揃えてください。現時点では Dev Container がもっとも再現性の高いセットアップ手段です。
