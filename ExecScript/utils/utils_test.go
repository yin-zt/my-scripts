package utils

import (
	"testing"
)

func TestExecScript(t *testing.T) {
	testCases := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:    "valid script",
			script:  "echo 'Hello World'",
			wantErr: false,
		},
		{
			name:    "invalid command",
			script:  "invalid_command_xyz",
			wantErr: true,
		},
		{
			name:    "empty script",
			script:  "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ExecScript(tc.script) // 现在可以直接调用同包函数

			if (err != nil) != tc.wantErr {
				t.Errorf("ExecScript() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr && len(output) == 0 {
				t.Error("Expected non-empty output for valid script")
			}
		})
	}
}
