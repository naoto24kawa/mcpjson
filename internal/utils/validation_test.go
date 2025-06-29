package utils

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		resourceType string
		wantErr      bool
	}{
		{"valid name", "test-profile", "プロファイル", false},
		{"valid with underscore", "test_profile", "プロファイル", false},
		{"valid alphanumeric", "profile123", "プロファイル", false},
		{"empty name", "", "プロファイル", true},
		{"too long", "a123456789012345678901234567890123456789012345678901", "プロファイル", true},
		{"invalid chars", "test@profile", "プロファイル", true},
		{"reserved word", "help", "プロファイル", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input, tt.resourceType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			"valid single env",
			"PORT=3000",
			map[string]string{"PORT": "3000"},
			false,
		},
		{
			"valid multiple env",
			"PORT=3000,DEBUG=true,API_KEY=secret",
			map[string]string{"PORT": "3000", "DEBUG": "true", "API_KEY": "secret"},
			false,
		},
		{
			"empty string",
			"",
			nil,
			false,
		},
		{
			"invalid format",
			"PORT:3000",
			nil,
			true,
		},
		{
			"invalid key",
			"123PORT=3000",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvVars(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != len(tt.want) {
				t.Errorf("ParseEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single arg", "arg1", []string{"arg1"}},
		{"multiple args", "arg1,arg2,arg3", []string{"arg1", "arg2", "arg3"}},
		{"args with spaces", "arg1, arg2 , arg3", []string{"arg1", "arg2", "arg3"}},
		{"empty string", "", nil},
		{"empty args", ",,", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgs(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("ParseArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
