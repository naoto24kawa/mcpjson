package utils

import (
	"io"
	"os"
	"testing"
)

func TestParseProfileName(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		defaultName    string
		expectedName   string
		expectedOffset int
		expectedOutput string
	}{
		{
			name:           "引数なし",
			args:           []string{},
			defaultName:    "default",
			expectedName:   "default",
			expectedOffset: 0,
			expectedOutput: "プロファイル名が指定されていないため、デフォルト 'default' を使用します\n",
		},
		{
			name:           "プロファイル名指定",
			args:           []string{"myprofile"},
			defaultName:    "default",
			expectedName:   "myprofile",
			expectedOffset: 1,
			expectedOutput: "",
		},
		{
			name:           "最初の引数がオプション",
			args:           []string{"--from", "path/to/file"},
			defaultName:    "default",
			expectedName:   "default",
			expectedOffset: 0,
			expectedOutput: "プロファイル名が指定されていないため、デフォルト 'default' を使用します\n",
		},
		{
			name:           "最初の引数がショートオプション",
			args:           []string{"-f", "path/to/file"},
			defaultName:    "default",
			expectedName:   "default",
			expectedOffset: 0,
			expectedOutput: "プロファイル名が指定されていないため、デフォルト 'default' を使用します\n",
		},
		{
			name:           "プロファイル名とオプション",
			args:           []string{"myprofile", "--from", "path/to/file"},
			defaultName:    "default",
			expectedName:   "myprofile",
			expectedOffset: 1,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			profileName, argsOffset := ParseProfileName(tt.args, tt.defaultName)

			// Restore stdout and read the output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)

			if profileName != tt.expectedName {
				t.Errorf("プロファイル名が期待値と異なります: got %q, want %q", profileName, tt.expectedName)
			}

			if argsOffset != tt.expectedOffset {
				t.Errorf("引数オフセットが期待値と異なります: got %d, want %d", argsOffset, tt.expectedOffset)
			}

			if string(output) != tt.expectedOutput {
				t.Errorf("出力が期待値と異なります: got %q, want %q", string(output), tt.expectedOutput)
			}
		})
	}
}

func TestParseFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		index         int
		flag          string
		expectedValue string
		expectedIndex int
		expectError   bool
	}{
		{
			name:          "正常なフラグ解析",
			args:          []string{"--from", "value"},
			index:         0,
			flag:          "--from",
			expectedValue: "value",
			expectedIndex: 1,
			expectError:   false,
		},
		{
			name:        "値が不足",
			args:        []string{"--from"},
			index:       0,
			flag:        "--from",
			expectError: true,
		},
		{
			name:        "範囲外のインデックス",
			args:        []string{"--from"},
			index:       1,
			flag:        "--from",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, index, err := ParseFlag(tt.args, tt.index, tt.flag)

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが返されませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}

				if value != tt.expectedValue {
					t.Errorf("値が期待値と異なります: got %q, want %q", value, tt.expectedValue)
				}

				if index != tt.expectedIndex {
					t.Errorf("インデックスが期待値と異なります: got %d, want %d", index, tt.expectedIndex)
				}
			}
		})
	}
}

func TestParseRenameArgs(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		defaultName        string
		expectedOldName    string
		expectedNewName    string
		expectedArgsOffset int
		expectError        bool
		expectedOutput     string
	}{
		{
			name:               "新旧両方指定",
			args:               []string{"old", "new"},
			defaultName:        "default",
			expectedOldName:    "old",
			expectedNewName:    "new",
			expectedArgsOffset: 2,
			expectError:        false,
			expectedOutput:     "",
		},
		{
			name:               "新しい名前のみ（旧名前はデフォルト）",
			args:               []string{"new"},
			defaultName:        "default",
			expectedOldName:    "default",
			expectedNewName:    "new",
			expectedArgsOffset: 1,
			expectError:        false,
			expectedOutput:     "元のプロファイル名が指定されていないため、デフォルト 'default' を使用します\n",
		},
		{
			name:        "引数なし",
			args:        []string{},
			defaultName: "default",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			oldName, newName, argsOffset, err := ParseRenameArgs(tt.args, tt.defaultName)

			// Restore stdout and read the output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが返されませんでした")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}

				if oldName != tt.expectedOldName {
					t.Errorf("旧名前が期待値と異なります: got %q, want %q", oldName, tt.expectedOldName)
				}

				if newName != tt.expectedNewName {
					t.Errorf("新名前が期待値と異なります: got %q, want %q", newName, tt.expectedNewName)
				}

				if argsOffset != tt.expectedArgsOffset {
					t.Errorf("引数オフセットが期待値と異なります: got %d, want %d", argsOffset, tt.expectedArgsOffset)
				}

				if string(output) != tt.expectedOutput {
					t.Errorf("出力が期待値と異なります: got %q, want %q", string(output), tt.expectedOutput)
				}
			}
		})
	}
}