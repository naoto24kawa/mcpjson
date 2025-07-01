# mcpjson

MCP（Model Context Protocol）設定ファイルを効率的に管理するCLIツール

## 概要

mcpjsonは、MCPサーバー設定をプロファイルとして管理し、異なる環境や用途に応じて簡単に切り替えることができるツールです。プロファイル機能により、複数のMCPサーバー設定を1つのプロファイルとして管理し、環境に応じた設定切り替えを実現します。

### 主な機能

#### 🔧 プロファイル管理
複数のMCPサーバー設定を1つのプロファイルとして管理

- プロファイルの作成・保存・削除
- 既存のMCP設定ファイルからプロファイル作成
- プロファイルの一覧表示・詳細表示
- 任意のパスへのプロファイル適用

#### 🖥️ MCPサーバー管理
MCPサーバー設定をテンプレートとして再利用

- サーバー設定をテンプレートとして保存
- サーバーの手動作成・編集
- サーバーの一覧表示・削除
- プロファイルへのサーバー追加・削除
- MCP設定ファイルからのサーバー情報抽出

## インストール

### Homebrew（推奨）

macOSユーザーの場合、Homebrewを使用したインストールが最も簡単です：

```bash
# tap を追加してインストール
brew tap naoto24kawa/mcpjson
brew install mcpjson

# または1行で
brew install naoto24kawa/mcpjson
```

### インストールスクリプト

最も簡単な方法は、インストールスクリプトを使用することです：

```bash
curl -sSL https://raw.githubusercontent.com/naoto24kawa/mcpjson/main/install.sh | bash
```

### バイナリダウンロード

[リリースページ](https://github.com/naoto24kawa/mcpjson/releases)から、お使いのOS/アーキテクチャに対応したバイナリをダウンロードしてください。

```bash
# Linux/macOS
curl -L https://github.com/naoto24kawa/mcpjson/releases/latest/download/mcpjson-linux-amd64 -o mcpjson
chmod +x mcpjson
sudo mv mcpjson /usr/local/bin/

# Windows
# https://github.com/naoto24kawa/mcpjson/releases から mcpjson-windows-amd64.exe をダウンロード
```

### Go install

Go 1.21以上がインストールされている場合：

```bash
go install github.com/naoto24kawa/mcpjson@latest
```

### ソースからビルド

開発者向け：

```bash
git clone https://github.com/naoto24kawa/mcpjson.git
cd mcpjson
go build -o mcpjson .
sudo mv mcpjson /usr/local/bin/  # Linux/macOS
```

## 使い方

### 🔰 初心者ガイド

#### 基本的な確認から始める

```bash
# 現在利用可能なプロファイルを確認
mcpjson list

# 現在のMCP設定ファイルの内容を確認
mcpjson server show --from ~/.mcp.json
```

#### プロファイル基本操作

```bash
# 1. 空のプロファイルを新規作成
mcpjson create my-profile

# 2. 既存のMCP設定ファイルからプロファイルを保存
mcpjson save work-profile --from ~/.mcp.json

# 3. プロファイルを適用
mcpjson apply work-profile --to /path/to/new/.mcp.json

# 4. プロファイル名を変更
mcpjson rename old-profile new-profile

# 5. 不要なプロファイルを削除
mcpjson delete old-profile
```

### サーバーテンプレート管理

#### 基本操作

```bash
# サーバー一覧を確認
mcpjson server list

# 既存のMCP設定ファイルからサーバーを抽出・保存
mcpjson server save git-server --server git --from ~/.mcp.json

# 手動でシンプルなサーバーを作成
mcpjson server save nodejs-server --command "node" --args "server.js,--port,3000"

# 環境変数を含むサーバーを作成
mcpjson server save api-server --command "python" --args "app.py" --env "PORT=8080,DEBUG=true"
```

#### 高度な環境変数管理

```bash
# 環境変数ファイルを使用してサーバーを作成
mcpjson server save prod-server --command "node" --args "server.js" --env-file ".env.production"

# 環境変数ファイル + 個別指定（個別指定が優先）
mcpjson server save dev-server --command "node" --env-file ".env.production" --env "DEBUG=true,PORT=4000"

# サーバーの部分更新
mcpjson server save prod-server --command "python"  # コマンドのみ更新
mcpjson server save prod-server --args ""            # 引数を削除
```

### プロファイルとサーバーの連携

```bash
# サーバーをMCPファイルに追加
mcpjson server add git-server --to ~/.mcp.json

# 環境変数をオーバーライドして追加
mcpjson server add nodejs-server --to ~/.mcp.json --as my-node --env "PORT=4000,DEBUG=false"

# MCPファイルからサーバーを削除
mcpjson server remove git --from ~/.mcp.json
```

### 実用的なワークフロー例

#### 開発環境から本番環境への移行

```bash
# 1. 現在の開発環境設定を確認
mcpjson server show --from .mcp.json

# 2. 開発用プロファイルを作成（バックアップ）
mcpjson save dev-profile

# 3. 本番用プロファイルを作成し、環境変数を調整
mcpjson create prod-profile
mcpjson server add git-server --env "GIT_REPO_PATH=/prod/repo,GIT_AUTHOR_EMAIL=prod@company.com"
mcpjson server add database-server --env "DB_HOST=prod-db.company.com,DB_SSL=true"

# 4. 本番環境に適用
mcpjson apply prod-profile --to /etc/claude/.mcp.json

# 5. 開発環境に戻すときは
mcpjson apply dev-profile
```

#### チームでの設定共有と個人カスタマイズ

```bash
# 1. チーム共通の標準サーバーを作成（チームリーダーが実行）
mcpjson server save team-git --command "uvx" --args "mcp-server-git,--repository,PROJECT_ROOT" --env "GIT_AUTHOR_NAME=TEAM_MEMBER"
mcpjson server save team-fs --command "uvx" --args "mcp-server-filesystem,--allowed-dirs,PROJECT_ROOT"

# 2. 個人用プロファイルを作成（各メンバーが実行）
mcpjson create my-profile
mcpjson server add team-git --to ~/.mcp.json --env "GIT_AUTHOR_NAME=Alice Johnson,PROJECT_ROOT=/Users/alice/work"
mcpjson server add team-fs --to ~/.mcp.json --env "PROJECT_ROOT=/Users/alice/work"

# 3. 個人環境に適用
mcpjson apply my-profile
```

## コマンドリファレンス

### 基本構文

```bash
mcpjson <コマンド> [オプション] [引数]
```

### プロファイル管理

| コマンド | 説明 | 例 |
|---------|------|-----|
| `apply [名前] --to <パス>` | プロファイルを指定パスに適用 | `mcpjson apply work-profile --to ~/.mcp.json` |
| `save [名前] --from <パス>` | 現在の設定をプロファイルとして保存 | `mcpjson save work-profile --from ~/.mcp.json` |
| `create [名前]` | 新規プロファイルを作成 | `mcpjson create my-profile` |
| `list [--detail]` | プロファイル一覧を表示 | `mcpjson list --detail` |
| `delete [名前] [--force]` | プロファイルを削除 | `mcpjson delete old-profile` |
| `rename [現在名] <新名前>` | プロファイル名を変更 | `mcpjson rename old new` |

### サーバー管理

#### サーバー保存・作成

```bash
# 設定ファイルから抽出
mcpjson server save <サーバー名> --server <サーバー名> --from <設定ファイルパス>

# 手動作成
mcpjson server save <サーバー名> --command <コマンド> [--args <引数>] [--env <環境変数>] [--env-file <ファイル>]
```

#### その他のサーバー操作

| コマンド | 説明 | 例 |
|---------|------|-----|
| `server list [--detail]` | テンプレート一覧を表示 | `mcpjson server list --detail` |
| `server delete <名前>` | テンプレートを削除 | `mcpjson server delete old-server` |
| `server rename <現在名> <新名前>` | テンプレート名を変更 | `mcpjson server rename old new` |
| `server add <テンプレート> --to <ファイル>` | MCPファイルにサーバー追加 | `mcpjson server add git-server --to ~/.mcp.json` |
| `server remove <サーバー名> --from <ファイル>` | MCPファイルからサーバー削除 | `mcpjson server remove git --from ~/.mcp.json` |
| `server show --from <ファイル>` | 設定ファイルのサーバー情報を表示 | `mcpjson server show --from ~/.mcp.json` |

### ユーティリティコマンド

| コマンド | 説明 | 例 |
|---------|------|-----|
| `detail <名前>` | プロファイルの詳細をJSON形式で表示 | `mcpjson detail work-profile` |
| `detail server <名前>` | サーバーテンプレートの詳細をJSON形式で表示 | `mcpjson detail server git-server` |
| `path [名前]` | プロファイルファイルの絶対パスを表示 | `mcpjson path work-profile` |
| `server-path <名前>` | サーバーテンプレートファイルの絶対パスを表示 | `mcpjson server-path git-server` |
| `reset <all\|profiles\|servers>` | 開発用設定のリセット | `mcpjson reset all --force` |

### プロファイル名のデフォルト値

プロファイル名を省略した場合、`default` が自動的に使用されます。

**対象コマンド:** `apply`, `save`, `create`, `delete`, `rename`

```bash
# 以下のコマンドは同等です
mcpjson apply --to ~/.mcp.json
mcpjson apply default --to ~/.mcp.json
```

### オプション詳細

#### 環境変数の指定

```bash
# 単一の環境変数
--env "PORT=3000"

# 複数の環境変数（カンマ区切り）
--env "PORT=3000,DEBUG=true,HOST=localhost"

# 環境変数ファイルの使用
--env-file ".env.production"

# 環境変数ファイル + 個別指定（個別指定が優先）
--env-file ".env" --env "DEBUG=true"
```

#### 引数の指定

```bash
# 単一の引数
--args "server.js"

# 複数の引数（カンマ区切り）
--args "server.js,--port,3000,--verbose"
```

## 設定ファイルの場所

mcpjsonは以下のディレクトリに設定を保存します：

```
~/.mcpjson/
├── profiles/     # プロファイル（.jsonc形式）
└── servers/      # サーバーテンプレート（.jsonc形式）
```

### ファイル形式

| ファイル種別 | 形式 | 説明 |
|------------|------|------|
| プロファイル | JSONC | 使用するサーバーの参照リスト（コメント付きJSON） |
| MCPサーバー | JSONC | 個別サーバー設定のテンプレート（コメント付きJSON） |
| MCP設定ファイル | JSON | `.mcp.json`等のMCP設定ファイル |

## トラブルシューティング

### よくあるエラーと解決方法

#### 存在しないプロファイルエラー

```bash
# ❌ エラー例
mcpjson apply nonexistent-profile
# エラー: プロファイル 'nonexistent-profile' が見つかりません

# ✅ 解決方法
mcpjson list  # 利用可能なプロファイルを確認
```

#### 環境変数の形式エラー

```bash
# ❌ 間違った形式
mcpjson server save myserver --env "PORT:3000"  # : を使用

# ✅ 正しい形式
mcpjson server save myserver --env "PORT=3000,DEBUG=true"  # = を使用
```

#### 権限不足エラー

```bash
# ❌ 権限エラー
mcpjson apply work-profile --to /etc/claude/.mcp.json

# ✅ 解決方法
sudo mcpjson apply work-profile --to /etc/claude/.mcp.json
# または書き込み可能なパスを使用
mcpjson apply work-profile --to ~/claude/.mcp.json
```

## 技術仕様

### 実行環境

- **対応OS**: Windows, macOS, Linux
- **実装言語**: Go
- **Goバージョン**: 1.21以上
- **依存関係**: なし（シングルバイナリとして配布）


## 開発者向け情報

mcpjsonの開発に参加したい方は、[DEVELOPER.md](DEVELOPER.md)をご覧ください。

- 開発環境の構築方法
- ビルド・テスト手順
- リリースプロセス
- コントリビューション方法

## ライセンス

MIT License