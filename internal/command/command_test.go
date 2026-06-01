package command_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/command"
)

func TestExecAppenderCreateCommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		appender command.ExecAppender
		args     []string
		wantArgs []string
		wantEnv  []string
	}{
		{
			name: "base args only with no appended args",
			appender: command.ExecAppender{
				Name: "git",
				Args: []string{"diff"},
			},
			args:     nil,
			wantArgs: []string{"git", "diff"},
			wantEnv:  nil,
		},
		{
			name: "base args with appended args",
			appender: command.ExecAppender{
				Name: "git",
				Args: []string{"diff", "--cached"},
			},
			args:     []string{"path/to/file"},
			wantArgs: []string{"git", "diff", "--cached", "path/to/file"},
			wantEnv:  nil,
		},
		{
			name: "no base args with appended args",
			appender: command.ExecAppender{
				Name: "git",
			},
			args:     []string{"status"},
			wantArgs: []string{"git", "status"},
			wantEnv:  nil,
		},
		{
			name: "env propagated to command",
			appender: command.ExecAppender{
				Name: "git",
				Args: []string{"diff"},
				Env:  []string{"FOO=bar", "BAZ=qux"},
			},
			args:     nil,
			wantArgs: []string{"git", "diff"},
			wantEnv:  []string{"FOO=bar", "BAZ=qux"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			opts := []cmp.Option{cmpopts.EquateEmpty()}

			// Act
			created := tc.appender.CreateCommand(ctx, tc.args...)

			// Assert
			execCmd, ok := created.(*exec.Cmd)
			if !ok {
				t.Fatalf("ExecAppender.CreateCommand(...) type = %T, want *exec.Cmd", created)
			}
			if got, want := execCmd.Args, tc.wantArgs; !cmp.Equal(got, want, opts...) {
				t.Errorf("ExecAppender.CreateCommand(...) Args = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
			}
			if got, want := execCmd.Env, tc.wantEnv; !cmp.Equal(got, want, opts...) {
				t.Errorf("ExecAppender.CreateCommand(...) Env = mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
			}
		})
	}
}
