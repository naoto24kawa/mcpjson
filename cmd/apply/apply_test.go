package apply

import (
	"os"
	"testing"

	"github.com/naoto24kawa/mcpjson/internal/config"
	"github.com/naoto24kawa/mcpjson/internal/profile"
	"github.com/naoto24kawa/mcpjson/internal/server"
	"github.com/naoto24kawa/mcpjson/internal/testutil"
)

func setupTestEnvironment(t *testing.T) (string, *config.Config, func()) {
	t.Helper()
	return testutil.SetupIsolatedTestEnvironment(t)
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setup      func(cfg *config.Config) string
		wantExit   bool
		useProfile bool
	}{
		{
			name:       "デフォルトプロファイルの適用",
			args:       []string{},
			useProfile: false,
			setup: func(cfg *config.Config) string {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				serverName := testutil.GenerateUniqueServerName("test-server")
				instanceName := testutil.GenerateUniqueServerName("my-test-server")

				_ = serverManager.SaveManual(serverName, "python", []string{"test.py"}, nil, false)
				_ = profileManager.Create("default", "")
				_ = profileManager.AddServer("default", serverName, instanceName, nil)
				return "default"
			},
			wantExit: false,
		},
		{
			name:       "特定プロファイルの適用",
			args:       []string{}, // Will be set dynamically in the test
			useProfile: true,
			setup: func(cfg *config.Config) string {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				profileName := testutil.GenerateUniqueProfileName("test-profile")
				serverName := testutil.GenerateUniqueServerName("test-server")
				instanceName := testutil.GenerateUniqueServerName("my-test-server")

				_ = serverManager.SaveManual(serverName, "python", []string{"test.py"}, nil, false)
				_ = profileManager.Create(profileName, "")
				_ = profileManager.AddServer(profileName, serverName, instanceName, nil)
				return profileName
			},
			wantExit: false,
		},
		{
			name:       "カスタムパスでの適用",
			args:       []string{}, // Will be set dynamically in the test
			useProfile: false,
			setup: func(cfg *config.Config) string {
				profileManager := profile.NewManager(cfg.ProfilesDir)
				serverManager := server.NewManager(cfg.ServersDir)

				serverName := testutil.GenerateUniqueServerName("test-server")
				instanceName := testutil.GenerateUniqueServerName("my-test-server")

				_ = serverManager.SaveManual(serverName, "python", []string{"test.py"}, nil, false)
				_ = profileManager.Create("test-profile-custom", "")
				_ = profileManager.AddServer("test-profile-custom", serverName, instanceName, nil)
				return "test-profile-custom"
			},
			wantExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cfg, cleanup := setupTestEnvironment(t)
			defer cleanup()

			// Setup test data
			profileName := tt.setup(cfg)

			// Handle dynamic args for different test cases
			args := tt.args
			if tt.useProfile {
				args = []string{profileName}
			} else if tt.name == "カスタムパスでの適用" {
				customPath := testutil.CreateTempFile(t, cfg.ProfilesDir, "custom-config", ".json", []byte("{}"))
				args = []string{profileName, "--to", customPath}
			}

			Execute(args)

			var outputPath string
			if len(args) >= 3 && (args[len(args)-2] == "--to" || args[len(args)-2] == "-t") {
				outputPath = args[len(args)-1]
			} else {
				outputPath = config.GetDefaultMCPPath()
			}

			if outputPath != config.GetDefaultMCPPath() {
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("出力ファイルが作成されませんでした: %s", outputPath)
				}
			}
		})
	}
}
