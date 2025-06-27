package server

import (
	"fmt"

	"github.com/naoto24kawa/mcpconfig/internal/interaction"
)

// TemplateUpdater handles template update operations
type TemplateUpdater struct {
	templateManager *TemplateManager
}

// NewTemplateUpdater creates a new TemplateUpdater instance
func NewTemplateUpdater(templateManager *TemplateManager) *TemplateUpdater {
	return &TemplateUpdater{
		templateManager: templateManager,
	}
}

// SaveManual saves or updates a server template manually
func (tu *TemplateUpdater) SaveManual(templateName, command string, args []string, env map[string]string, force bool) error {
	existing := tu.templateExists(templateName)

	if existing && !force {
		if !interaction.ConfirmOverwrite("サーバーテンプレート", templateName) {
			return fmt.Errorf("上書きをキャンセルしました")
		}
	}

	var template *ServerTemplate
	var err error

	if existing {
		template, err = tu.updateExistingTemplate(templateName, command, args, env)
		if err != nil {
			return err
		}
		fmt.Printf("サーバーテンプレート '%s' を更新しました\n", templateName)
	} else {
		template, err = tu.createNewTemplate(templateName, command, args, env)
		if err != nil {
			return err
		}
		fmt.Printf("サーバーテンプレート '%s' を作成しました\n", templateName)
	}

	return tu.templateManager.save(template)
}

func (tu *TemplateUpdater) updateExistingTemplate(templateName, command string, args []string, env map[string]string) (*ServerTemplate, error) {
	template, err := tu.templateManager.Load(templateName)
	if err != nil {
		return nil, err
	}

	if command != "" {
		template.ServerConfig.Command = command
	}

	tu.UpdateTemplateArgs(template, args)
	tu.UpdateTemplateEnv(template, env)

	return template, nil
}

func (tu *TemplateUpdater) createNewTemplate(templateName, command string, args []string, env map[string]string) (*ServerTemplate, error) {
	if command == "" {
		return nil, fmt.Errorf("コマンドが指定されていません")
	}

	return CreateServerTemplate(templateName, command, args, env), nil
}

func (tu *TemplateUpdater) UpdateTemplateArgs(template *ServerTemplate, args []string) {
	if args == nil {
		return
	}

	if len(args) == 1 && args[0] == "" {
		template.ServerConfig.Args = nil
	} else {
		template.ServerConfig.Args = args
	}
}

func (tu *TemplateUpdater) UpdateTemplateEnv(template *ServerTemplate, env map[string]string) {
	if env == nil {
		return
	}

	if len(env) == 0 {
		template.ServerConfig.Env = nil
		return
	}

	if template.ServerConfig.Env == nil {
		template.ServerConfig.Env = make(map[string]string)
	}

	for k, v := range env {
		if v == "" {
			delete(template.ServerConfig.Env, k)
		} else {
			template.ServerConfig.Env[k] = v
		}
	}
}

func (tu *TemplateUpdater) templateExists(name string) bool {
	exists, _ := tu.templateManager.Exists(name)
	return exists
}