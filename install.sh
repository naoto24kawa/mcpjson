#!/bin/bash

set -e

# カラー定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 設定
REPO="naoto24kawa/mcpconfig"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="mcpconfig"

# ヘルパー関数
print_error() {
    echo -e "${RED}エラー: $1${NC}" >&2
}

print_success() {
    echo -e "${GREEN}$1${NC}"
}

print_info() {
    echo -e "${YELLOW}$1${NC}"
}

# OS/アーキテクチャの検出
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $OS in
        darwin) OS="darwin" ;;
        linux) OS="linux" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *)
            print_error "サポートされていないOS: $OS"
            exit 1
            ;;
    esac
    echo $OS
}

detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64|amd64) ARCH="x86_64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            print_error "サポートされていないアーキテクチャ: $ARCH"
            exit 1
            ;;
    esac
    echo $ARCH
}

# 最新リリースの取得
get_latest_release() {
    local release_url="https://api.github.com/repos/$REPO/releases/latest"
    local latest_version=$(curl -s $release_url | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$latest_version" ]; then
        print_error "最新バージョンの取得に失敗しました"
        exit 1
    fi
    
    echo $latest_version
}

# ダウンロードとインストール
download_and_install() {
    local version=$1
    local os=$2
    local arch=$3
    
    # ファイル名の構築
    local filename="${BINARY_NAME}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
        filename="${filename}.zip"
    else
        filename="${filename}.tar.gz"
    fi
    
    local download_url="https://github.com/$REPO/releases/download/$version/$filename"
    local temp_dir=$(mktemp -d)
    
    print_info "ダウンロード中: $download_url"
    
    # ダウンロード
    if ! curl -L -o "$temp_dir/$filename" "$download_url"; then
        print_error "ダウンロードに失敗しました"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # 展開
    cd "$temp_dir"
    if [ "$os" = "windows" ]; then
        unzip -q "$filename"
    else
        tar -xzf "$filename"
    fi
    
    # インストール
    if [ "$os" = "windows" ]; then
        print_info "Windowsでは手動でPATHに追加してください: $temp_dir/${BINARY_NAME}.exe"
    else
        # sudoが必要かチェック
        if [ -w "$INSTALL_DIR" ]; then
            mv "$BINARY_NAME" "$INSTALL_DIR/"
        else
            print_info "管理者権限が必要です..."
            sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
        fi
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # クリーンアップ
    cd - > /dev/null
    rm -rf "$temp_dir"
}

# メイン処理
main() {
    print_info "mcpconfig インストーラー"
    print_info "========================"
    
    # OS/アーキテクチャの検出
    OS=$(detect_os)
    ARCH=$(detect_arch)
    print_info "検出されたシステム: $OS/$ARCH"
    
    # 最新バージョンの取得
    VERSION=$(get_latest_release)
    print_info "最新バージョン: $VERSION"
    
    # ダウンロードとインストール
    download_and_install "$VERSION" "$OS" "$ARCH"
    
    # 確認
    if command -v $BINARY_NAME &> /dev/null; then
        print_success "インストールが完了しました!"
        print_info "バージョン: $($BINARY_NAME --version)"
    else
        print_error "インストールに失敗しました"
        exit 1
    fi
}

# 実行
main "$@"