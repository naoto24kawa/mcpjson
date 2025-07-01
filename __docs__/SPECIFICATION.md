# mcpjson 仕様書

## 概要

mcpjsonは、MCP（Model Context Protocol）設定ファイルを効率的に管理するCLIツールです。  
プロファイル機能により、異なる環境や用途に応じたMCPサーバー設定を簡単に切り替えできます。

## 主要機能

### 🔧 プロファイル管理
複数のMCPサーバー設定を1つのプロファイルとして管理

- プロファイルの作成・保存・削除
- 既存のMCP設定ファイルからプロファイル作成
- プロファイルの一覧表示・詳細表示
- 任意のパスへのプロファイル適用

### 🖥️ MCPサーバー管理
MCPサーバー設定をサーバーとして再利用

- サーバー設定をサーバーとして保存
- サーバーの手動作成・編集
- サーバーの一覧表示・削除
- プロファイルへのサーバー追加・削除
- MCP設定ファイルからのサーバー情報抽出

## 技術仕様

### 💻 実行環境
- **対応OS**: Windows, macOS, Linux  
- **実装言語**: Go
- **Goバージョン**: 1.21以上

### 📦 依存関係
- なし（シングルバイナリとして配布）
- Go標準ライブラリのみ使用

### 📁 ファイル構成
```
mcpjson/
├── main.go           # メインエントリーポイント
├── go.mod            # Go modules定義
├── go.sum            # 依存関係チェックサム
├── cmd/
│   └── root.go       # CLIルートコマンド
├── internal/
│   ├── config/       # 設定管理
│   ├── profile/      # プロファイル操作
│   ├── server/       # MCPサーバー操作
│   └── utils/        # ユーティリティ関数
└── README.md
```

### 🚀 インストール方法

#### バイナリダウンロード（推奨）
```bash
# GitHubリリースから最新版をダウンロード
# Linux/macOS
curl -L https://github.com/your-org/mcpjson/releases/latest/download/mcpjson-linux-amd64 -o mcpjson
chmod +x mcpjson
sudo mv mcpjson /usr/local/bin/

# Windows
# https://github.com/your-org/mcpjson/releases から mcpjson-windows-amd64.exe をダウンロード
```

#### ソースからビルド
```bash
# Go 1.21以上が必要
git clone https://github.com/your-org/mcpjson.git
cd mcpjson
go build -o mcpjson .
sudo mv mcpjson /usr/local/bin/  # Linux/macOS
```

## CLI仕様

### 📝 基本構文
```bash
mcpjson <コマンド> [オプション] [引数]
```

### 🔗 プロファイル名のデフォルト値
プロファイル名を指定する各コマンドでは、プロファイル名を省略した場合に **`default`** が自動的に使用されます。

**対象コマンド:**
- `apply`, `save`, `create`, `delete`, `rename`

**注意:** `server add`, `server remove`コマンドは、MCPファイルパスを指定しない場合に **`./.mcp.json`** がデフォルトで使用されます。

**例:**
```bash
# 以下のコマンドは同等です
mcpjson apply --to ~/.mcp.json
mcpjson apply default --to ~/.mcp.json
```

### 📋 コマンド一覧

#### `apply` - プロファイルを指定パスに適用
```bash
mcpjson apply [プロファイル名] [--to <適用先パス>]
mcpjson apply [プロファイル名] [-t <適用先パス>]
```

**パラメータ:**
- `[プロファイル名]` - 適用するプロファイル名（省略時: `default`）
- `--to, -t` - 適用先のMCP設定ファイルパス（省略時: `./.mcp.json`）

> プロファイル内のサーバー参照を解決してMCP設定ファイルを生成します

#### `save` - 現在の設定をプロファイルとして保存
```bash
mcpjson save [プロファイル名] [--from <設定ファイルパス>] [--force]
mcpjson save [プロファイル名] [-f <設定ファイルパス>] [-F]
```

**パラメータ:**
- `[プロファイル名]` - 保存するプロファイル名（省略時: `default`）
- `--from, -f` - 保存元のMCP設定ファイルパス（省略時: `./.mcp.json`）
- `--force, -F` - 既存プロファイルを確認なしで上書き

> MCP設定ファイル内の各サーバーは、対応するサーバーが存在する場合はサーバー参照として保存し、存在しない場合は新しいサーバーを自動作成します

##### 自動サーバー抽出の仕組み
プロファイル保存時(`save`コマンド)に以下の処理を実行：

1. **サーバー名の決定**
   - MCP設定ファイル内のサーバー名をサーバー名として使用
   - 例：設定内の`"git"`サーバー → `git.json`サーバー

2. **重複処理**
   - 既存サーバーが存在する場合は既存を優先（先勝）
   - ユーザーに通知メッセージを表示：`"サーバー 'git' は既に存在するため、既存のものを使用します"`

3. **サーバー作成**
   - 新規サーバーの`description`フィールドは`null`に設定
   - サーバー設定（command, args, env, timeout, envFile, transportType）をそのまま抽出してサーバー化

#### `create` - 新規プロファイルを作成
```bash
mcpjson create [プロファイル名] [--template <サーバー名>]
mcpjson create [プロファイル名] [-t <サーバー名>]
```

**パラメータ:**
- `[プロファイル名]` - 作成するプロファイル名（省略時: `default`）
- `--template, -t` - 使用するサーバー（省略時は空のプロファイル）

#### `list` - プロファイル一覧を表示
```bash
mcpjson list [--detail]
mcpjson list [-d]
```

**オプション:**
- `--detail, -d` - 詳細情報も表示

#### `delete` - プロファイルを削除
```bash
mcpjson delete [プロファイル名] [--force]
mcpjson delete [プロファイル名] [-f]
```

**パラメータ:**
- `[プロファイル名]` - 削除するプロファイル名（省略時: `default`）
- `--force, -f` - 確認なしで削除

#### `rename` - プロファイル名を変更
```bash
mcpjson rename [現在のプロファイル名] <新しいプロファイル名> [--force]
mcpjson rename [現在のプロファイル名] <新しいプロファイル名> [-f]
```

**パラメータ:**
- `[現在のプロファイル名]` - 変更前のプロファイル名（省略時: `default`）
- `<新しいプロファイル名>` - 変更後のプロファイル名
- `--force, -f` - 既存プロファイルがある場合に確認なしで上書き

#### `copy` - プロファイルをコピー
```bash
mcpjson copy [コピー元プロファイル名] <コピー先プロファイル名> [--force]
mcpjson copy [コピー元プロファイル名] <コピー先プロファイル名> [-f]
```

**パラメータ:**
- `[コピー元プロファイル名]` - コピー元のプロファイル名（省略時: `default`）
- `<コピー先プロファイル名>` - コピー先のプロファイル名
- `--force, -f` - 既存プロファイルがある場合に確認なしで上書き

> 既存のプロファイルを別名でコピーします。元のプロファイルはそのまま残ります

#### `merge` - 複数のプロファイルを合成
```bash
mcpjson merge <合成先プロファイル名> <ソースプロファイル1> [ソースプロファイル2] ... [--force]
mcpjson merge <合成先プロファイル名> <ソースプロファイル1> [ソースプロファイル2] ... [-f]
```

**パラメータ:**
- `<合成先プロファイル名>` - 合成後のプロファイル名
- `<ソースプロファイル1>` - 合成元のプロファイル名（必須）
- `[ソースプロファイル2] ...` - 追加の合成元プロファイル名（省略可）
- `--force, -f` - 既存プロファイルがある場合に確認なしで上書き

> 複数のプロファイルのサーバー設定を1つのプロファイルに統合します。重複するサーバー名は先勝ちで処理されます

#### `detail` - 詳細をJSON形式で表示
```bash
mcpjson detail <プロファイル名>       # プロファイルの詳細を表示
```

**パラメータ:**
- `<プロファイル名>` - 詳細を表示するプロファイル名

> プロファイルの全データをJSONC形式で出力します

#### `path` - プロファイルパス表示
```bash
mcpjson path [プロファイル名]
```

**パラメータ:**
- `[プロファイル名]` - パスを表示するプロファイル名（省略時: `default`）

> 指定したプロファイルファイルの絶対パスを表示します

#### `server-path` - サーバーテンプレートパス表示
```bash
mcpjson server-path <サーバーテンプレート名>
```

**パラメータ:**
- `<サーバーテンプレート名>` - パスを表示するサーバーテンプレート名

> 指定したサーバーテンプレートファイルの絶対パスを表示します

#### `reset` - 開発用設定のリセット
```bash
mcpjson reset <サブコマンド> [オプション]
```

**サブコマンド:**
- `all` - すべての設定をリセット（プロファイル + サーバーテンプレート）
- `profiles` - すべてのプロファイルを削除
- `servers` - すべてのサーバーテンプレートを削除

**オプション:**
- `--force, -f` - 確認なしで実行

> 開発・テスト用として、保存された設定を一括でリセットします

#### `server` - MCPサーバー管理

**サーバー保存（設定ファイルから）**
```bash
mcpjson server save <サーバー名> --server <サーバー名> --from <設定ファイルパス> [--force]
mcpjson server save <サーバー名> -s <サーバー名> -f <設定ファイルパス> [-F]
```

**サーバー手動作成**
```bash
mcpjson server save <サーバー名> --command <コマンド> [--args <引数1,引数2>] [--env <KEY=VALUE,KEY2=VALUE2>] [--env-file <環境変数ファイルパス>] [--timeout <秒数>] [--transport-type <タイプ>] [--force]
mcpjson server save <サーバー名> -c <コマンド> [-a <引数1,引数2>] [-e <KEY=VALUE,KEY2=VALUE2>] [--env-file <環境変数ファイルパス>] [--timeout <秒数>] [--transport-type <タイプ>] [-F]
```

**サーバー一覧表示**
```bash
mcpjson server list [--detail]
mcpjson server list [-d]
```

**サーバー削除**
```bash
mcpjson server delete <サーバー名> [--force]
mcpjson server delete <サーバー名> [-f]
```

**サーバー名変更**
```bash
mcpjson server rename <現在のサーバー名> <新しいサーバー名> [--force]
mcpjson server rename <現在のサーバー名> <新しいサーバー名> [-f]
```

**MCPファイルにサーバー追加**
```bash
mcpjson server add <サーバーテンプレート名> [--to <MCPファイルパス>] [--as <サーバー名>] [--env <KEY=VALUE,KEY2=VALUE2>]
mcpjson server add <サーバーテンプレート名> [-t <MCPファイルパス>] [-a <サーバー名>] [-e <KEY=VALUE,KEY2=VALUE2>]
```

**MCPファイルからサーバー削除**
```bash
mcpjson server remove <サーバー名> [--from <MCPファイルパス>]
mcpjson server remove <サーバー名> [-f <MCPファイルパス>]
```


##### パラメータ詳細

**設定ファイルからサーバー保存時:**
- `<サーバー名>` - 保存するサーバー名
- `--server, -s` - 抽出対象のサーバー名
- `--from, -f` - MCP設定ファイルパス
- `--force, -F` - 既存サーバーを確認なしで上書き

**手動でサーバー作成時:**
- `<サーバー名>` - 保存するサーバー名
- `--command, -c` - 実行コマンド（必須、新規作成時）
- `--args, -a` - 引数（カンマ区切り、空文字で削除）
- `--env, -e` - 環境変数（KEY=VALUE形式、カンマ区切り、空文字で削除）
- `--env-file` - 環境変数ファイルパス（.envファイル形式）
- `--timeout` - タイムアウト秒数（数値）
- `--transport-type` - 通信タイプ（例：stdio）
- `--force, -F` - 既存サーバーを確認なしで上書き

> 既存サーバー更新時：オプション指定なしは現在値維持、空文字指定で削除  
> `--env`と`--env-file`の併用時：ファイルの環境変数を先に読み込み、その後`--env`の値で上書き

**その他のパラメータ:**
- `list --detail, -d` - 詳細情報も表示
- `delete --force, -f` - 確認なしで削除
- `rename --force, -f` - 既存サーバーがある場合に確認なしで上書き
- `add --to, -t` - 追加先MCPファイルパス（省略時: `./.mcp.json`）
- `add --as, -a` - 追加時のサーバー名（省略時はサーバーテンプレート名）
- `add --env, -e` - 環境変数オーバーライド
- `remove --from, -f` - 削除元MCPファイルパス（省略時: `./.mcp.json`）

### 🌐 グローバルオプション
- `--help, -h` - ヘルプを表示
- `--version, -v` - バージョンを表示

### 🛠️ ユーティリティコマンド
以下のコマンドは開発・運用支援用のユーティリティ機能です。

**パス表示コマンド:**
- `path` - プロファイルファイルの絶対パスを表示
- `server-path` - サーバーテンプレートファイルの絶対パスを表示

**リセットコマンド:**
- `reset` - 開発用設定の一括リセット

### 📝 使用例

> 🔰 **初心者ガイド**: 以下の例は基本操作から高度な操作まで、学習しやすい順序で紹介しています。  
> 初めて使用する方は「🔰 はじめに」から始めて、順番に進んでください。

#### 🔰 はじめに - 基本的な確認
```bash
# 現在利用可能なプロファイルを確認
mcpjson list
# 出力例:
# プロファイル名        作成日時               サーバー数
# work-profile        2024-01-01 10:00:00    3
# dev-profile         2024-01-01 11:00:00    2

# 現在のMCP設定ファイルの内容を確認
mcpjson server list --detail --from ~/.mcp.json
```

#### 📁 プロファイル基本操作
```bash
# 1. 空のプロファイルを新規作成（デフォルトプロファイル使用）
mcpjson create
# 出力: プロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を作成しました

# 2. 空のプロファイルを新規作成（名前指定）
mcpjson create my-profile
# 出力: プロファイル 'my-profile' を作成しました

# 3. 既存のMCP設定ファイルからプロファイルを保存（デフォルトプロファイル、カレントディレクトリのmcp.json使用）
mcpjson save
# 出力: プロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を保存しました (3個のサーバー)
# 　　　サーバー 'git' を作成しました
# 　　　サーバー 'filesystem' を作成しました

# 4. 既存のMCP設定ファイルからプロファイルを保存（名前指定）
mcpjson save work-profile --from ~/.mcp.json
# 出力: プロファイル 'work-profile' を保存しました (3個のサーバー)

# 5. プロファイルを適用（デフォルトプロファイル使用、カレントディレクトリのmcp.json）
mcpjson apply
# 出力: プロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を適用しました
# 　　　3個のサーバー設定を './.mcp.json' に保存

# 6. プロファイルを別の場所に適用（名前指定）
mcpjson apply work-profile --to /path/to/new/.mcp.json
# 出力: プロファイル 'work-profile' を適用しました

# 7. プロファイル名を変更（デフォルトプロファイルから）
mcpjson rename new-profile
# 出力: 元のプロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を 'new-profile' に変更しました

# 8. プロファイル名を変更（名前指定）
mcpjson rename old-profile new-profile
# 出力: プロファイル 'old-profile' を 'new-profile' に変更しました

# 9. プロファイルをコピー（デフォルトプロファイルから）
mcpjson copy backup-profile
# 出力: プロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を 'backup-profile' にコピーしました

# 10. プロファイルをコピー（名前指定）
mcpjson copy work-profile work-profile-backup
# 出力: プロファイル 'work-profile' を 'work-profile-backup' にコピーしました

# 11. 複数のプロファイルを合成
mcpjson merge combined-profile dev-profile test-profile
# 出力: プロファイル 'combined-profile' を作成しました（5個のサーバー）
# 　　　合成元: [dev-profile test-profile]

# 12. 重複するサーバーがある場合の合成
mcpjson merge all-profile profile1 profile2 profile3
# 出力: 警告: サーバー 'git' は既に追加されているため、スキップします（プロファイル: profile2）
# 　　　プロファイル 'all-profile' を作成しました（7個のサーバー）
# 　　　合成元: [profile1 profile2 profile3]

# 13. 不要なプロファイルを削除（デフォルトプロファイル）
mcpjson delete
# 出力: プロファイル名が指定されていないため、デフォルト 'default' を使用します
# 　　　プロファイル 'default' を削除しました

# 14. 不要なプロファイルを削除（名前指定）
mcpjson delete old-profile
# 出力: プロファイル 'old-profile' を削除しました
```

#### 🛠️ サーバー基本操作
```bash
# 1. サーバー一覧を確認
mcpjson server list
# 出力例:
# サーバー名           作成日時               コマンド
# git-server          2024-01-01 10:00:00   uvx mcp-server-git
# nodejs-server       2024-01-01 11:00:00   node server.js

# 2. 既存のMCP設定ファイルからサーバーを抽出・保存
mcpjson server save git-server --server git --from ~/.mcp.json
# 出力: サーバー 'git-server' を保存しました
# 　　　コマンド: uvx mcp-server-git
# 　　　引数: ["--repository", "/Users/name/project"]

# 3. 手動でシンプルなサーバーを作成
mcpjson server save nodejs-server --command "node" --args "server.js,--port,3000"
# 出力: サーバー 'nodejs-server' を作成しました

# 4. 環境変数を含むサーバーを作成
mcpjson server save api-server --command "python" --args "app.py" --env "PORT=8080,DEBUG=true"
# 出力: サーバー 'api-server' を作成しました
# 　　　環境変数: PORT=8080, DEBUG=true
```

#### 🔧 環境変数ファイルを使った高度な操作
```bash
# .env.production ファイルの内容例:
# PORT=3000
# DEBUG=false
# DATABASE_URL=postgres://prod-server/db
# API_KEY=prod-api-key-12345

# 環境変数ファイルを使用してサーバーを作成
mcpjson server save prod-server --command "node" --args "server.js" --env-file ".env.production"
# 出力: サーバー 'prod-server' を作成しました
# 　　　環境変数ファイル '.env.production' から4個の環境変数を読み込み

# 環境変数ファイル + 個別指定（個別指定が優先）
mcpjson server save dev-server --command "node" --env-file ".env.production" --env "DEBUG=true,PORT=4000"
# 出力: サーバー 'dev-server' を作成しました
# 　　　最終的な環境変数: PORT=4000, DEBUG=true, DATABASE_URL=postgres://prod-server/db, API_KEY=prod-api-key-12345

# サーバーの部分更新（既存サーバーの一部のみ変更）
mcpjson server save prod-server --command "python"  # コマンドのみ更新
# 出力: サーバー 'prod-server' を更新しました（コマンドを変更）

mcpjson server save prod-server --args ""            # 引数を削除
# 出力: サーバー 'prod-server' を更新しました（引数を削除）

# サーバー名を変更
mcpjson server rename old-template new-template
# 出力: サーバー 'old-template' を 'new-template' に変更しました

# 不要なサーバーを削除
mcpjson server delete custom-server
# 出力: サーバー 'custom-server' を削除しました
```

#### MCPファイルとサーバーの連携
```bash
# サーバーをデフォルトMCPファイルに追加
mcpjson server add git-server
# 出力: MCPファイルパスが指定されていないため、デフォルト './.mcp.json' を使用します

# サーバーを指定MCPファイルに追加
mcpjson server add git-server --to ~/.mcp.json

# 環境変数をオーバーライドして追加
mcpjson server add nodejs-server --to ~/.mcp.json --as my-node --env "PORT=4000,DEBUG=false"

# デフォルトMCPファイルからサーバーを削除
mcpjson server remove git
# 出力: MCPファイルパスが指定されていないため、デフォルト './.mcp.json' を使用します

# 指定MCPファイルからサーバーを削除
mcpjson server remove git --from ~/.mcp.json
```

#### 🔍 設定ファイルの確認とデバッグ
```bash
# MCP設定ファイルの全サーバー情報を表示
mcpjson server list --detail --from ~/.mcp.json
```

#### 🚀 実用的なワークフロー例

**シナリオ1: 開発環境から本番環境への移行**
```bash
# 1. 現在の開発環境設定を確認
mcpjson server list --detail --from .mcp.json
# 目的: 何が設定されているか把握

# 2. 開発用プロファイルを作成（バックアップ）
mcpjson save dev-profile
# 出力: プロファイル 'dev-profile' を保存しました (3個のサーバー)

# 3. 本番用プロファイルを作成し、環境変数を調整
mcpjson create prod-profile
mcpjson server add git-server --to prod-profile --env "GIT_REPO_PATH=/prod/repo,GIT_AUTHOR_EMAIL=prod@company.com"
mcpjson server add database-server --to prod-profile --env "DB_HOST=prod-db.company.com,DB_SSL=true"
# 目的: 本番環境用のパラメータで設定をカスタマイズ

# 4. 本番環境に適用
mcpjson apply prod-profile --to /etc/claude/.mcp.json
# 出力: プロファイル 'prod-profile' を適用しました

# 5. 開発環境に戻すときは
mcpjson apply dev-profile
# 出力: プロファイル 'dev-profile' を適用しました
```

**シナリオ2: チームでの設定共有と個人カスタマイズ**
```bash
# 1. チーム共通の標準サーバーを作成（チームリーダーが実行）
mcpjson server save team-git --command "uvx" --args "mcp-server-git,--repository,PROJECT_ROOT" --env "GIT_AUTHOR_NAME=TEAM_MEMBER"
mcpjson server save team-fs --command "uvx" --args "mcp-server-filesystem,--allowed-dirs,PROJECT_ROOT"
# 目的: チーム全体で使える共通サーバーを準備

# 2. 個人用プロファイルを作成（各メンバーが実行）
mcpjson create my-profile
# 注意: 以下はMCPファイルに直接追加する例です
mcpjson server add team-git --to ~/.mcp.json --env "GIT_AUTHOR_NAME=Alice Johnson,PROJECT_ROOT=/Users/alice/work"
mcpjson server add team-fs --to ~/.mcp.json --env "PROJECT_ROOT=/Users/alice/work"
# 目的: 共通サーバーをベースに個人の環境に合わせてカスタマイズ

# 3. 個人環境に適用
mcpjson apply my-profile
# 成功: チーム標準 + 個人カスタマイズされた設定が適用される
```

#### 🛠️ ユーティリティコマンドの使用例

**パス表示:**
```bash
# プロファイルファイルのパスを表示
mcpjson path work-profile
# 出力: /Users/name/.mcpjson/profiles/work-profile.jsonc

# デフォルトプロファイルのパスを表示
mcpjson path
# 出力: /Users/name/.mcpjson/profiles/default.jsonc

# サーバーテンプレートのパスを表示
mcpjson server-path git-server
# 出力: /Users/name/.mcpjson/servers/git-server.jsonc
```

**リセット操作:**
```bash
# すべての設定をリセット（確認あり）
mcpjson reset all
# 出力: 以下の設定がすべて削除されます:
#       - すべてのプロファイル
#       - すべてのサーバーテンプレート
#       本当にすべての設定をリセットしますか？ (Y/n):

# プロファイルのみリセット（確認なし）
mcpjson reset profiles --force
# 出力: すべてのプロファイルをリセットしました

# サーバーテンプレートのみリセット
mcpjson reset servers
# 出力: 本当にすべてのサーバーテンプレートを削除しますか？ (Y/n):
```

#### ⚠️ トラブルシューティングガイド

**ケース1: 存在しないプロファイルエラー**
```bash
# ❌ 問題のあるコマンド:
mcpjson apply nonexistent-profile
# エラー: プロファイル 'nonexistent-profile' が見つかりません

# ✅ 解決方法:
mcpjson list  # 利用可能なプロファイルを確認
# 出力で正しいプロファイル名を確認して再実行
```

**ケース2: 環境変数の形式エラー**
```bash
# ❌ 問題のあるコマンド:
mcpjson server save myserver --command "node" --env "PORT:3000"  # ：を使用
# エラー: 環境変数の形式が不正です: 'PORT:3000'

# ✅ 正しいコマンド:
mcpjson server save myserver --command "node" --env "PORT=3000,DEBUG=true"
# 注意: = を使用し、複数の場合はカンマで区切る
```

**ケース3: 権限不足エラー**
```bash
# ❌ 問題のあるコマンド:
mcpjson apply work-profile --to /etc/claude/.mcp.json
# エラー: ファイル書き込み権限がありません

# ✅ 解決方法:
sudo mcpjson apply work-profile --to /etc/claude/.mcp.json  # sudoで実行
# または書き込み可能なパスを使用:
mcpjson apply work-profile --to ~/claude/.mcp.json
# またはカレントディレクトリを使用:
mcpjson apply work-profile  # ./.mcp.jsonに保存
```

**ケース4: 環境変数ファイルエラー**
```bash
# ❌ 問題のあるコマンド:
mcpjson server save api --command "python" --env-file ".env.missing"
# エラー: 環境変数ファイルが見つかりません: '.env.missing'

# ✅ 解決方法:
ls -la .env*  # 利用可能な.envファイルを確認
# 正しいファイルパスを指定して再実行
```

**ケース5: JSON形式エラー**
```bash
# ❌ 問題のある状況:
mcpjson save broken-profile --from ~/.mcp.json
# エラー: MCP設定ファイルのJSON形式が不正です

# ✅ 解決方法:
jq . ~/.mcp.json  # jqコマンドでJSON構文を検証
# エラー箱所を修正して再実行
```

## 📊 データ仕様

### 📁 ファイル形式
| ファイル種別 | 形式 | 説明 |
|------------|------|------|
| プロファイル | JSONC | 使用するサーバーの参照リスト（コメント付きJSON） |
| MCPサーバー | JSONC | 個別サーバー設定のサーバー（コメント付きJSON） |
| MCP設定ファイル | JSON | `.mcp.json`等のMCP設定ファイル |

**文字エンコーディング:** UTF-8

**JSONC形式について:**
- プロファイルとサーバーテンプレートでは、説明のためのコメントが記述可能
- 行コメント（`//`）とブロックコメント（`/* */`）をサポート
- 設定の意図や用途を明確化するために活用

### 📏 保存場所
```
~/.mcpjson/
├── profiles/              # プロファイル保存ディレクトリ
│   ├── work-profile.jsonc
│   ├── dev-profile.jsonc
│   └── ...
└── servers/               # MCPサーバー保存ディレクトリ
    ├── git-server.jsonc
    ├── nodejs-server.jsonc
    └── ...
```

**MCP設定ファイル:** ユーザー指定の任意のパス（例: `~/.mcp.json`）

### 📋 プロファイルJSONC形式
```jsonc
{
  // プロファイルメタデータ
  "name": "work-profile",
  "description": "作業用のMCPサーバー構成",
  "createdAt": "2024-01-01T00:00:00.000Z",
  "updatedAt": "2024-01-01T00:00:00.000Z",
  
  // 使用するサーバーの一覧
  "servers": [
    {
      "name": "git-tools",        // MCP設定ファイル内でのサーバー名
      "template": "git-server",   // 使用するサーバーテンプレート名
      "overrides": {
        "env": {
          // プロジェクト固有のリポジトリパス
          "GIT_REPO_PATH": "/path/to/my/repo"
        }
      }
    },
    {
      "name": "file-manager",     // ファイル操作用サーバー
      "template": "nodejs-server" // Node.jsベースのサーバーテンプレート
    }
  ]
}
```

> プロファイルは使用するサーバーの参照リストを保存し、apply時にMCP設定ファイルを生成します

### 🔧 MCPサーバーJSONC形式
```jsonc
{
  // サーバーテンプレートメタデータ
  "name": "git-server",
  "description": "Git操作用MCPサーバー",
  "createdAt": "2024-01-01T00:00:00.000Z",
  
  // サーバー実行設定
  "serverConfig": {
    "command": "uvx",  // 実行コマンド
    "args": [
      "mcp-server-git", 
      "--repository", 
      "/path/to/repo"   // デフォルトリポジトリパス（環境変数で上書き可能）
    ],
    "env": {
      "GIT_AUTHOR_NAME": "Your Name",         // Git作成者名
      "GIT_AUTHOR_EMAIL": "your.email@example.com"  // Git作成者メール
    },
    "timeout": 60,        // サーバー応答タイムアウト（秒）
    "envFile": "${workspaceFolder}/.claude/.env.local.mcp",  // 環境変数ファイル
    "transportType": "stdio"  // 通信方式
  }
}
```

> MCPサーバーは個別のサーバー設定（command, args, env, timeout, envFile, transportType等）を保存したサーバーです

## ⚠️ エラーハンドリング・バリデーション仕様

### 🔢 終了コード
| コード | 種別 | 説明 |
|------|------|------|
| 0 | 正常 | 正常終了 |
| 1 | 一般エラー | 引数不正、ファイル操作失敗等 |
| 2 | リソースエラー | プロファイル・サーバーが存在しない |
| 3 | ファイルエラー | MCP設定ファイルが存在しない・読み込み不可 |
| 4 | フォーマットエラー | JSON形式が不正 |
| 5 | 環境エラー | サポートされていないOS/アーキテクチャ |
| 6 | サーバーエラー | サーバー名が見つからない（MCP設定ファイル内） |
| 7 | 引数エラー | 手動作成時の引数形式エラー |
| 8 | 参照エラー | 参照先サーバーが見つからない（apply時） |

### ✅ バリデーション項目

#### 📝 プロファイル名・サーバー名
- **使用可能文字:** 英数字、ハイフン、アンダースコア (`[a-zA-Z0-9_-]+`)
- **最大長:** 50文字
- **予約語:** `help`, `version`, `list`, `server`等のコマンド名は禁止

#### 📁 ファイルパス
- **読み込み時:** 存在チェック
- **保存時:** 書き込み権限チェック
- **ディレクトリ:** 必要に応じて自動作成

#### 📜 JSON形式
- **構文チェック:** Go標準ライブラリで検証
- **必須フィールド:** `servers`, `serverConfig`
- **参照整合性:** サーバー参照の妥当性確認
- **重複チェック:** プロファイル内のサーバー名
- **サーバー存在:** apply時のサーバー名照合

#### 🔧 MCPサーバー手動作成
- **コマンド:** `--command`は新規作成時必須（空文字不可）
- **引数形式:** `--args` はカンマ区切り文字列（例: `"arg1,arg2,--option,value"`）
- **環境変数形式:** `--env` はKEY=VALUE形式をカンマ区切り（例: `"PORT=3000,DEBUG=true"`）
- **環境変数ファイル形式:** `--env-file` は.env形式のファイルパス（例: `.env`, `.env.production`）
- **タイムアウト:** `--timeout` は秒数（数値）（例: `60`）
- **通信タイプ:** `--transport-type` は通信方式（例: `stdio`）
- **KEY制約:** 環境変数のKEYは英数字とアンダースコアのみ
- **更新ルール:** 指定なしは現在値維持、空文字指定で削除
- **併用ルール:** `--env`と`--env-file`併用時は、ファイルを先に読み込み、`--env`で上書き

### 💬 エラーメッセージ例

#### 📝 引数エラー
```
エラー: プロファイル名が指定されていません
使用方法: mcpjson apply <プロファイル名> --to <パス>
```

#### 📁 ファイルエラー
```
エラー: プロファイル 'work-profile' が見つかりません
利用可能なプロファイル: mcpjson list
```

#### 📜 JSONエラー
```
エラー: MCP設定ファイルのJSON形式が不正です
ファイル: ~/.mcp.json
詳細: [Go JSONパースエラー]
```

#### 📦 実行環境エラー
```
エラー: サポートされていないOSまたはアーキテクチャです
対応環境: Windows/Linux/macOS (amd64/arm64)
```

#### 🔧 サーバー関連エラー
```
エラー: MCPサーバー 'git' がMCP設定ファイルに見つかりません
ファイル: ~/.mcp.json
利用可能なサーバー: [サーバー名一覧]
```

```
エラー: サーバー名 'git' は既にプロファイル 'work-profile' に存在します
別の名前を指定してください: --as <新しい名前>
```

#### ⚙️ 手動作成時の引数エラー
```
エラー: コマンドが指定されていません
使用方法: mcpjson server save <サーバー名> --command <コマンド>
```

```
エラー: 環境変数の形式が不正です: 'INVALID_FORMAT'
正しい形式: KEY=VALUE （例: PORT=3000）
```

```
エラー: 環境変数ファイルが見つかりません: '.env.production'
ファイルパスを確認してください
```

```
エラー: 環境変数ファイルの読み込みに失敗しました: '.env'
詳細: [ファイル読み込みエラー]
```

```
エラー: 環境変数ファイルの形式が不正です: '.env' 行3
正しい形式: KEY=VALUE （コメント行は#で開始）
```

#### ⚠️ 上書き確認
```
警告: プロファイル 'work-profile' は既に存在します
上書きしますか？ (y/N): 
```

#### 📝 名前変更関連エラー
```
エラー: プロファイル 'old-profile' が見つかりません
利用可能なプロファイル: mcpjson list
```

```
エラー: プロファイル 'new-profile' は既に存在します
別の名前を指定するか、--force オプションで上書きしてください
```

```
エラー: サーバー 'old-template' が見つかりません
利用可能なサーバー: mcpjson server list
```

```
エラー: サーバー 'new-template' は既に存在します
別の名前を指定するか、--force オプションで上書きしてください
```

```
警告: サーバー 'old-template' はプロファイル 'work-profile' で参照されています
名前を変更するとプロファイルの参照が無効になります。継続しますか？ (y/N):
```

### 📢 エラー出力方針
- **出力先:** 標準エラー出力 (stderr)
- **言語:** 日本語でわかりやすいメッセージ
- **解決ヒント:** エラー解決のための情報を含める
- **詳細情報:** `--verbose`オプション時のみ表示

### 🔄 上書き確認の動作
- **確認プロンプト:** 既存ファイルがある場合に表示
- **強制上書き:** `--force`オプションで確認をスキップ
- **入力判定:** `y` (Yes)以外は全てキャンセル扱い
- **非対話環境:** パイプ等では自動的にキャンセル

---

## 📋 開発ノート

この仕様書は、MCPサーバー設定の効率的な管理を目的としたCLIツール「mcpjson」の完全な技術仕様です。

**設計方針:**
- シンプルかつ直感的なCLI操作
- プロファイル機能による設定の使い回し
- サーバー機能による個別サーバー設定の再利用
- 堅牢なエラーハンドリングとバリデーション

**実装時の注意点:**
- 機能の過剰実装は避け、必要最小限の機能に留める
- JSONファイルの整合性とバリデーションを重視
- ユーザビリティを考慮したエラーメッセージの提供
