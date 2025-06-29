package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/naoto24kawa/mcpconfig/internal/config"
)

// TemplateDisplay handles template listing and display operations
type TemplateDisplay struct {
	serversDir string
}

// NewTemplateDisplay creates a new TemplateDisplay instance
func NewTemplateDisplay(serversDir string) *TemplateDisplay {
	return &TemplateDisplay{
		serversDir: serversDir,
	}
}

// List displays all server templates
func (td *TemplateDisplay) List(detail bool) error {
	files, err := os.ReadDir(td.serversDir)
	if err != nil {
		return fmt.Errorf("サーバーディレクトリの読み込みに失敗しました: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("サーバーテンプレートが存在しません")
		return nil
	}

	if detail {
		return td.listDetailed(files)
	}

	return td.listSummary(files)
}

func (td *TemplateDisplay) listDetailed(files []os.DirEntry) error {
	templateManager := NewTemplateManager(td.serversDir)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			name := strings.TrimSuffix(file.Name(), config.FileExtension)
			template, err := templateManager.Load(name)
			if err != nil {
				fmt.Printf("エラー: %s の読み込みに失敗しました: %v\n", name, err)
				continue
			}

			td.displayTemplateDetail(template)
		}
	}
	return nil
}

func (td *TemplateDisplay) listSummary(files []os.DirEntry) error {
	templateManager := NewTemplateManager(td.serversDir)

	fmt.Printf("%-20s %-20s %s\n", "テンプレート名", "作成日時", "コマンド")
	fmt.Println(strings.Repeat("-", 60))

	for _, file := range files {
		if strings.HasSuffix(file.Name(), config.FileExtension) {
			name := strings.TrimSuffix(file.Name(), config.FileExtension)
			template, err := templateManager.Load(name)
			if err != nil {
				continue
			}

			fmt.Printf("%-*s %-*s %s\n",
				ListColumnWidth, template.Name,
				ListColumnWidth, template.CreatedAt.Format(TimestampFormat),
				template.ServerConfig.Command)
		}
	}
	return nil
}

func (td *TemplateDisplay) displayTemplateDetail(template *ServerTemplate) {
	fmt.Printf("\nテンプレート: %s\n", template.Name)
	if template.Description != nil {
		fmt.Printf("  説明: %s\n", *template.Description)
	}
	fmt.Printf("  作成日時: %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  コマンド: %s\n", template.ServerConfig.Command)
	if len(template.ServerConfig.Args) > 0 {
		fmt.Printf("  引数: %v\n", template.ServerConfig.Args)
	}
	if len(template.ServerConfig.Env) > 0 {
		fmt.Println("  環境変数:")
		for k, v := range template.ServerConfig.Env {
			fmt.Printf("    %s=%s\n", k, v)
		}
	}
}
