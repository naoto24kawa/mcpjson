package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	rnd "math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/naoto24kawa/mcpconfig/internal/config"
)

// UniqueTestID generates a unique identifier for tests
func UniqueTestID() string {
	// Create a random seed source
	source := rnd.NewSource(time.Now().UnixNano())
	r := rnd.New(source)
	
	// Generate random bytes
	bytes := make([]byte, 4)
	for i := range bytes {
		bytes[i] = byte(r.Intn(256))
	}
	
	return hex.EncodeToString(bytes)
}

// GenerateUniqueProfileName generates a unique profile name for testing
func GenerateUniqueProfileName(baseName string) string {
	if baseName == "" {
		baseName = "test-profile"
	}
	return fmt.Sprintf("%s-%s", baseName, UniqueTestID())
}

// GenerateUniqueServerName generates a unique server name for testing
func GenerateUniqueServerName(baseName string) string {
	if baseName == "" {
		baseName = "test-server"
	}
	return fmt.Sprintf("%s-%s", baseName, UniqueTestID())
}

// SetupIsolatedTestEnvironment creates an isolated test environment with unique names
func SetupIsolatedTestEnvironment(t *testing.T) (tempDir string, cfg *config.Config, cleanup func()) {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mcpconfig-test-"+UniqueTestID())
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// Backup original environment variables
	origXDGConfigHome := os.Getenv("XDG_CONFIG_HOME")
	origHome := os.Getenv("HOME")

	// Set temporary environment
	_ = os.Setenv("XDG_CONFIG_HOME", tempDir)
	_ = os.Setenv("HOME", tempDir)

	cleanup = func() {
		// Restore original environment
		_ = os.Setenv("XDG_CONFIG_HOME", origXDGConfigHome)
		_ = os.Setenv("HOME", origHome)
		
		// Remove temporary directory
		_ = os.RemoveAll(tempDir)
	}

	// Create config
	cfg, err = config.New()
	if err != nil {
		cleanup()
		t.Fatalf("設定の作成に失敗: %v", err)
	}

	// Create necessary directories
	_ = os.MkdirAll(cfg.ProfilesDir, 0755)
	_ = os.MkdirAll(cfg.ServersDir, 0755)

	return tempDir, cfg, cleanup
}

// GenerateRandomID generates a random ID for tests
func GenerateRandomID(length int) string {
	if length <= 0 {
		length = 8
	}
	
	bytes := make([]byte, int(math.Ceil(float64(length)/2)))
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to math/rand if crypto/rand fails
		source := rnd.NewSource(time.Now().UnixNano())
		r := rnd.New(source)
		for i := range bytes {
			bytes[i] = byte(r.Intn(256))
		}
	}
	
	result := hex.EncodeToString(bytes)
	if len(result) > length {
		result = result[:length]
	}
	
	return result
}

// CreateTempFile creates a temporary file with unique name
func CreateTempFile(t *testing.T, dir, prefix, suffix string, content []byte) string {
	t.Helper()
	
	filename := fmt.Sprintf("%s-%s%s", prefix, UniqueTestID(), suffix)
	filepath := filepath.Join(dir, filename)
	
	if err := os.WriteFile(filepath, content, 0644); err != nil {
		t.Fatalf("一時ファイルの作成に失敗: %v", err)
	}
	
	return filepath
}