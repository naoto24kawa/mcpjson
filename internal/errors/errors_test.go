package errors

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	message := "バリデーションエラーです"
	err := NewValidationError(message)

	if err.Type != TypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, TypeValidation)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Code != 7 {
		t.Errorf("Code = %v, want %v", err.Code, 7)
	}
	if err.Cause != nil {
		t.Errorf("Cause = %v, want nil", err.Cause)
	}
}

func TestNewFileError(t *testing.T) {
	message := "ファイルエラーです"
	cause := errors.New("原因のエラー")
	err := NewFileError(message, cause)

	if err.Type != TypeFile {
		t.Errorf("Type = %v, want %v", err.Type, TypeFile)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Code != 3 {
		t.Errorf("Code = %v, want %v", err.Code, 3)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewConfigError(t *testing.T) {
	message := "設定エラーです"
	cause := errors.New("設定の原因エラー")
	err := NewConfigError(message, cause)

	if err.Type != TypeConfig {
		t.Errorf("Type = %v, want %v", err.Type, TypeConfig)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Code != 5 {
		t.Errorf("Code = %v, want %v", err.Code, 5)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewGeneralError(t *testing.T) {
	message := "一般的なエラーです"
	cause := errors.New("一般的な原因エラー")
	err := NewGeneralError(message, cause)

	if err.Type != TypeGeneral {
		t.Errorf("Type = %v, want %v", err.Type, TypeGeneral)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Code != 1 {
		t.Errorf("Code = %v, want %v", err.Code, 1)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "原因エラーありの場合",
			err: &AppError{
				Message: "メインエラー",
				Cause:   errors.New("原因エラー"),
			},
			expected: "メインエラー: 原因エラー",
		},
		{
			name: "原因エラーなしの場合",
			err: &AppError{
				Message: "メインエラー",
				Cause:   nil,
			},
			expected: "メインエラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("原因エラー")
	err := &AppError{
		Message: "メインエラー",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("AppError.Unwrap() = %v, want %v", unwrapped, cause)
	}

	// 原因エラーがない場合
	errNoCause := &AppError{
		Message: "メインエラー",
		Cause:   nil,
	}

	unwrappedNil := errNoCause.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("AppError.Unwrap() = %v, want nil", unwrappedNil)
	}
}
