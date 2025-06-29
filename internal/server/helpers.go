package server

import "time"

// CreateServerTemplate creates a new ServerTemplate instance
func CreateServerTemplate(name, command string, args []string, env map[string]string) *ServerTemplate {
	return &ServerTemplate{
		Name:        name,
		Description: nil,
		CreatedAt:   time.Now(),
		ServerConfig: ServerConfig{
			Command: command,
			Args:    args,
			Env:     env,
		},
	}
}
