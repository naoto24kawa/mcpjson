# mcpconfig

MCP（Model Context Protocol）設定ファイルを効率的に管理するCLIツール

## 概要

mcpconfigは、MCPサーバー設定をプロファイルとして管理し、異なる環境や用途に応じて簡単に切り替えることができるツールです。

### 主な機能

- **プロファイル管理**: 複数のMCPサーバー設定を1つのプロファイルとして管理
- **サーバーテンプレート**: MCPサーバー設定をテンプレートとして再利用
- **環境別設定**: 開発・本番環境など、用途に応じた設定の切り替え
- **設定の共有**: チームでの標準設定の共有と個人カスタマイズ

## インストール

### インストールスクリプト（推奨）

最も簡単な方法は、インストールスクリプトを使用することです：

```bash
curl -sSL https://raw.githubusercontent.com/naoto24kawa/mcpconfig/main/install.sh | bash
```

### バイナリダウンロード

[リリースページ](https://github.com/naoto24kawa/mcpconfig/releases)から、お使いのOS/アーキテクチャに対応したバイナリをダウンロードしてください。

```bash
# 例: macOS (Apple Silicon)
curl -L https://github.com/naoto24kawa/mcpconfig/releases/latest/download/mcpconfig_Darwin_arm64.tar.gz | tar xz
sudo mv mcpconfig /usr/local/bin/
```

### Go install

Go 1.21以上がインストールされている場合：

```bash
go install github.com/naoto24kawa/mcpconfig@latest
```

### ソースからビルド

開発者向け：

```bash
git clone https://github.com/naoto24kawa/mcpconfig.git
cd mcpconfig
make build  # または go build -o mcpconfig .
make install  # /usr/local/bin にインストール
```

## 使い方

### 基本的な使用例

#### 1. 現在のMCP設定をプロファイルとして保存

```bash
mcpconfig save work-profile --from ~/.mcp.json
```

#### 2. プロファイルを別の場所に適用

```bash
mcpconfig apply work-profile --path ~/projects/myapp/.mcp.json
```

#### 3. プロファイル一覧を確認

```bash
mcpconfig list
```

### サーバーテンプレート管理

#### テンプレートを手動で作成

```bash
mcpconfig server save git-server --command "uvx" --args "mcp-server-git,--repository,/path/to/repo"
```

#### 環境変数を含むテンプレート作成

```bash
mcpconfig server save api-server --command "python" --args "app.py" --env "PORT=8080,DEBUG=true"
```

#### プロファイルにテンプレートを追加

```bash
mcpconfig server add git-server --to work-profile
```

### 高度な使用例

#### 開発環境から本番環境への移行

```bash
# 開発環境の設定を保存
mcpconfig save dev-profile --from ~/.mcp.json

# 本番用プロファイルを作成
mcpconfig create prod-profile
mcpconfig server add git-server --to prod-profile --env "GIT_REPO_PATH=/prod/repo"
mcpconfig server add api-server --to prod-profile --env "PORT=3000,DEBUG=false"

# 本番環境に適用
mcpconfig apply prod-profile --path /etc/claude/.mcp.json
```

## コマンドリファレンス

### プロファイル管理

- `apply <name> --path <path>` - プロファイルを指定パスに適用
- `save <name> --from <path>` - 現在の設定をプロファイルとして保存
- `create <name>` - 新規プロファイルを作成
- `list [--detail]` - プロファイル一覧を表示
- `delete <name>` - プロファイルを削除
- `rename <old> <new>` - プロファイル名を変更

### サーバー管理

- `server save` - サーバーテンプレートを保存
- `server list` - テンプレート一覧を表示
- `server delete` - テンプレートを削除
- `server rename` - テンプレート名を変更
- `server add` - プロファイルにテンプレートを追加
- `server remove` - プロファイルからサーバーを削除
- `server show` - 設定ファイルのサーバー情報を表示

## 設定ファイルの場所

mcpconfigは以下のディレクトリに設定を保存します：

```
~/.mcpconfig/
├── profiles/     # プロファイル
└── servers/      # サーバーテンプレート
```

## 開発

### 必要な環境

- Go 1.21以上
- Make（オプション）

### ビルド

```bash
# 通常のビルド
make build

# 全プラットフォーム向けビルド
make build-all

# テスト実行
make test

# リリース用スナップショット作成
make snapshot
```

### リリース

新しいバージョンをリリースする場合：

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actionsが自動的にバイナリをビルドし、リリースを作成します。

## ライセンス

MIT License