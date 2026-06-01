package commandtest_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/command"
	"rodusek.dev/pkg/cpp-tools/internal/command/commandtest"
)

var errTest = errors.New("test error")

func TestErrOnStart(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.ErrOnStart(errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	startErr := cmd.Start()

	// Assert
	if got, want := startErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnStart(...).CreateCommand(...).Start() = %v, want %v", got, want)
	}
}

func TestErrOnStderrPipe(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.ErrOnStderrPipe(errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	stderr, pipeErr := cmd.StderrPipe()

	// Assert
	if got, want := pipeErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnStderrPipe(...).CreateCommand(...).StderrPipe() error = %v, want %v", got, want)
	}
	if got, want := stderr, io.ReadCloser(nil); !cmp.Equal(got, want) {
		t.Errorf("ErrOnStderrPipe(...).CreateCommand(...).StderrPipe() stderr = %v, want %v", got, want)
	}
}

func TestErrOnStdoutPipe(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.ErrOnStdoutPipe(errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	stdout, pipeErr := cmd.StdoutPipe()

	// Assert
	if got, want := pipeErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnStdoutPipe(...).CreateCommand(...).StdoutPipe() error = %v, want %v", got, want)
	}
	if got, want := stdout, io.ReadCloser(nil); !cmp.Equal(got, want) {
		t.Errorf("ErrOnStdoutPipe(...).CreateCommand(...).StdoutPipe() stdout = %v, want %v", got, want)
	}
}

func TestErrOnWait(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.ErrOnWait(errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	startErr := cmd.Start()
	waitErr := cmd.Wait()

	// Assert
	if got, want := startErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnWait(...).CreateCommand(...).Start() = %v, want %v", got, want)
	}
	if got, want := waitErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnWait(...).CreateCommand(...).Wait() = %v, want %v", got, want)
	}
}

func TestErrOnPipe(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.ErrOnPipe(errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	_, stdoutErr := cmd.StdoutPipe()
	_, stderrErr := cmd.StderrPipe()

	// Assert
	if got, want := stdoutErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnPipe(...).CreateCommand(...).StdoutPipe() error = %v, want %v", got, want)
	}
	if got, want := stderrErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ErrOnPipe(...).CreateCommand(...).StderrPipe() error = %v, want %v", got, want)
	}
}

func TestPipes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		stdout     string
		stderr     string
		wantStdout string
		wantStderr string
	}{
		{
			name:       "empty pipes",
			stdout:     "",
			stderr:     "",
			wantStdout: "",
			wantStderr: "",
		},
		{
			name:       "populated pipes",
			stdout:     "out content",
			stderr:     "err content",
			wantStdout: "out content",
			wantStderr: "err content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			ctx := context.Background()
			sut := commandtest.Pipes(tc.stdout, tc.stderr)

			// Act
			cmd := sut.CreateCommand(ctx)
			stdoutContent, stderrContent, runErr := runAndRead(t, cmd)

			// Assert
			if got, want := runErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Pipes(...).CreateCommand(...) run error = %v, want %v", got, want)
			}
			if got, want := stdoutContent, tc.wantStdout; !cmp.Equal(got, want) {
				t.Errorf("Pipes(...) stdout = %q, want %q", got, want)
			}
			if got, want := stderrContent, tc.wantStderr; !cmp.Equal(got, want) {
				t.Errorf("Pipes(...) stderr = %q, want %q", got, want)
			}
		})
	}
}

func TestPipesAndWaitErr(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.PipesAndWaitErr("out", "err", errTest)

	// Act
	cmd := sut.CreateCommand(ctx)
	stdoutContent, stderrContent, runErr := runAndRead(t, cmd)

	// Assert
	if got, want := runErr, errTest; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("PipesAndWaitErr(...).CreateCommand(...) run error = %v, want %v", got, want)
	}
	if got, want := stdoutContent, "out"; !cmp.Equal(got, want) {
		t.Errorf("PipesAndWaitErr(...) stdout = %q, want %q", got, want)
	}
	if got, want := stderrContent, "err"; !cmp.Equal(got, want) {
		t.Errorf("PipesAndWaitErr(...) stderr = %q, want %q", got, want)
	}
}

func TestFakeCommandStartAfterStart(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.Pipes("", "")
	cmd := sut.CreateCommand(ctx)
	firstErr := cmd.Start()

	// Act
	secondErr := cmd.Start()

	// Assert
	if got, want := firstErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...) first Start() = %v, want %v", got, want)
	}
	if got, want := secondErr, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...) second Start() = %v, want non-nil error", got)
	}
}

func TestFakeCommandStderrPipeAfterStart(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.Pipes("", "")
	cmd := sut.CreateCommand(ctx)
	startErr := cmd.Start()

	// Act
	stderr, pipeErr := cmd.StderrPipe()

	// Assert
	if got, want := startErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...).Start() = %v, want %v", got, want)
	}
	if got, want := pipeErr, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...).StderrPipe() after Start error = %v, want non-nil error", got)
	}
	if got, want := stderr, io.ReadCloser(nil); !cmp.Equal(got, want) {
		t.Errorf("Pipes(...).StderrPipe() after Start stderr = %v, want %v", got, want)
	}
}

func TestFakeCommandStdoutPipeAfterStart(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.Pipes("", "")
	cmd := sut.CreateCommand(ctx)
	startErr := cmd.Start()

	// Act
	stdout, pipeErr := cmd.StdoutPipe()

	// Assert
	if got, want := startErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...).Start() = %v, want %v", got, want)
	}
	if got, want := pipeErr, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...).StdoutPipe() after Start error = %v, want non-nil error", got)
	}
	if got, want := stdout, io.ReadCloser(nil); !cmp.Equal(got, want) {
		t.Errorf("Pipes(...).StdoutPipe() after Start stdout = %v, want %v", got, want)
	}
}

func TestFakeCommandWaitBeforeStart(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	sut := commandtest.Pipes("", "")
	cmd := sut.CreateCommand(ctx)

	// Act
	waitErr := cmd.Wait()

	// Assert
	if got, want := waitErr, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Pipes(...).Wait() before Start = %v, want non-nil error", got)
	}
}

// runAndRead reads stdout and stderr from cmd and returns their contents along
// with the wait error.
func runAndRead(t *testing.T, cmd command.Command) (string, string, error) {
	t.Helper()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	if err := cmd.Start(); err != nil {
		return "", "", err
	}
	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		return "", "", err
	}
	errBytes, err := io.ReadAll(stderr)
	if err != nil {
		return "", "", err
	}
	if err := cmd.Wait(); err != nil {
		return string(outBytes), string(errBytes), err
	}
	return string(outBytes), string(errBytes), nil
}
