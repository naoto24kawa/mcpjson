# mcpjson 開発者ガイド

mcpjsonの開発に参加していただき、ありがとうございます。このドキュメントでは、開発環境の構築から実際の開発手順まで、コントリビューターに必要な情報を提供します。

## 必要な環境

### 基本要件

- **Go**: 1.21以上
- **Git**: バージョン管理
- **Make**: ビルドタスク管理（オプション）

### 推奨ツール

- **IDE**: VS Code, GoLand等のGo対応エディタ
- **ターミナル**: bash, zsh等

## 開発環境の構築

### 1. リポジトリのクローン

```bash
git clone https://github.com/naoto24kawa/mcpjson.git
cd mcpjson
```

### 2. 依存関係の確認

```bash
go mod tidy
```

### 3. ビルドの確認

```bash
go build -o mcpjson .
```

## 開発手順

### ソースからビルド

```bash
# ローカルビルド
go build -o mcpjson .

# インストール
sudo mv mcpjson /usr/local/bin/  # Linux/macOS
```

### 開発時のビルド

```bash
# 通常のビルド
make build

# 全プラットフォーム向けビルド
make build-all

# 開発時のクイックビルド
go build -o bin/mcpjson .
```

### テスト実行

```bash
# 全テスト実行
make test

# 特定パッケージのテスト
go test ./internal/...

# カバレッジ付きテスト
go test -cover ./...

# ベンチマークテスト
go test -bench=. ./...
```

### 開発時のテスト

```bash
# ウォッチモードでテスト（要: air等のツール）
air test

# 特定のテスト関数のみ実行
go test -run TestSpecificFunction ./internal/...
```

## プロジェクト構造

```
mcpjson/
├── cmd/                    # コマンドラインインターフェース
│   └── root.go
├── internal/               # 内部パッケージ
│   ├── config/            # 設定管理
│   ├── profile/           # プロファイル管理
│   ├── server/            # サーバー管理
│   └── util/              # ユーティリティ
├── pkg/                   # 公開パッケージ
├── test/                  # テストデータ・ヘルパー
├── docs/                  # ドキュメント
├── scripts/               # ビルド・デプロイスクリプト
├── Makefile              # ビルドタスク
├── go.mod                # Go modules
└── README.md             # ユーザー向けドキュメント
```

## 技術仕様

### 実行環境

- **対応OS**: Windows, macOS, Linux
- **アーキテクチャ**: amd64, arm64
- **実装言語**: Go 1.21+
- **依存関係**: 標準ライブラリのみ（シングルバイナリ配布）

### 設定ファイル形式

| ファイル種別 | 形式 | 説明 |
|------------|------|------|
| プロファイル | JSONC | 使用するサーバーの参照リスト（コメント付きJSON） |
| MCPサーバー | JSONC | 個別サーバー設定のテンプレート（コメント付きJSON） |
| MCP設定ファイル | JSON | `.mcp.json`等のMCP設定ファイル |

### エラーコード

| コード | 説明 | 対応方法 |
|------|------|----------|
| 0 | 正常終了 | - |
| 1 | 一般エラー（引数不正、ファイル操作失敗等） | 引数・ファイルパスの確認 |
| 2 | リソースエラー（プロファイル・サーバーが存在しない） | リソース名の確認、list コマンドで確認 |
| 3 | ファイルエラー（MCP設定ファイルが存在しない・読み込み不可） | ファイルの存在・権限確認 |
| 4 | フォーマットエラー（JSON形式が不正） | JSON形式の確認・修正 |
| 5 | 環境エラー（サポートされていないOS/アーキテクチャ） | 対応環境の確認 |

## リリース手順

### 1. バージョンタグの作成

```bash
# 新しいバージョンをタグ付け
git tag v1.0.0
git push origin v1.0.0
```

### 2. 自動リリース

GitHub Actionsが自動的に以下を実行します：

- 全プラットフォーム向けバイナリのビルド
- リリースページの作成
- アーティファクトのアップロード

### 3. リリース用スナップショット作成（ローカル）

```bash
# GoReleaserを使用したスナップショット
make snapshot

# 手動でのクロスコンパイル
GOOS=linux GOARCH=amd64 go build -o dist/mcpjson-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o dist/mcpjson-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o dist/mcpjson-windows-amd64.exe .
```

## コントリビューション方法

### 1. 開発の流れ

1. **Issue の確認**: 既存のIssueを確認し、重複がないかチェック
2. **ブランチ作成**: `feature/` または `fix/` プレフィックスでブランチ作成
3. **開発**: 機能実装・バグ修正
4. **テスト**: 新規テスト追加・全テスト実行
5. **Pull Request**: 詳細な説明とともにPR作成

### 2. コミットメッセージ規約

```bash
# 形式: type(scope): subject

# 例
feat(server): サーバーテンプレートの環境変数マージ機能を追加
fix(profile): プロファイル適用時のパス解決を修正
docs(readme): インストール手順を更新
test(util): 環境変数パースのテストケースを追加
```

### 3. コードスタイル

- **gofmt**: 標準のコードフォーマット
- **golint**: リント規則に従う
- **go vet**: 静的解析チェック
- **変数名**: キャメルケース、短縮形は避ける
- **コメント**: 公開関数・型には必須

### 4. テストの書き方

```go
// テストファイル例: internal/config/config_test.go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {
            name:  "正常なJSONファイル",
            input: `{"servers": {"test": {}}}`,
            want:  &Config{Servers: map[string]Server{"test": {}}},
        },
        // 追加のテストケース...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

### 5. Pull Request のガイドライン

#### PRタイトル

- 明確で簡潔な説明
- Issueがある場合は番号を含める（例: `fix #123: プロファイル削除時のエラー処理`）

#### PR説明

```markdown
## 概要
このPRは何をするものか、なぜ必要かを説明

## 変更内容
- 機能A を追加
- バグB を修正
- テストC を改善

## テスト
- [ ] 既存テストが全て通る
- [ ] 新しいテストを追加した
- [ ] 手動テストを実施した

## 注意事項
- 破壊的変更の有無
- 依存関係の変更
- その他の注意点
```

## 開発時のデバッグ

### ログ出力

```go
// デバッグ用のログ出力
import "log"

func debugLog(format string, args ...interface{}) {
    if os.Getenv("DEBUG") == "true" {
        log.Printf("[DEBUG] "+format, args...)
    }
}
```

### 実行例

```bash
# デバッグモードで実行
DEBUG=true ./mcpjson list

# 詳細なGoの実行ログ
GODEBUG=gctrace=1 ./mcpjson list
```

## よくある質問

### Q: 新しい機能を追加したいのですが、どこから始めればいいですか？

A: まず Issue を作成し、実装方針を相談してください。その後、feature ブランチを作成して開発を開始してください。

### Q: テストが進まないのですが？

A: `go test -v ./...` で詳細なテスト結果を確認し、失敗の原因を特定してください。テストデータの作成には `test/` ディレクトリを活用してください。

### Q: ビルドエラーが発生します

A: `go mod tidy` で依存関係を整理し、Go のバージョンが 1.21 以上であることを確認してください。

## 関連リンク

- [メインREADME](README.md) - ユーザー向けドキュメント
- [GitHub Issues](https://github.com/naoto24kawa/mcpjson/issues) - バグ報告・機能要求
- [Go Documentation](https://golang.org/doc/) - Go言語の公式ドキュメント
- [MCP仕様](https://spec.modelcontextprotocol.io/) - Model Context Protocol仕様

---

ご質問やご不明な点があれば、Issue を作成するか、既存の Issue にコメントしてください。