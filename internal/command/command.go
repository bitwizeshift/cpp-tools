package command

import (
	"context"
	"io"
	"os/exec"
)

// Command is an interface that abstracts the execution of an [exec.Cmd] so that
// it can be more testable.
type Command interface {
	Start() error
	StderrPipe() (io.ReadCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	Wait() error
}

// Creator is a factory for [Command]s. It is used to abstract away the
// creation of [Command]s, allowing for easier testing and flexibility in how
// commands are created and executed.
type Creator interface {
	CreateCommand(ctx context.Context, arg ...string) Command
}

// ExecAppender is a [Creator] that creates [Command]s by appending arguments to
// a base command. It also allows for setting environment variables for the
// created commands.
type ExecAppender struct {
	// Name is the base command to execute. It should be the name of an executable
	// available in the system's PATH or an absolute path to an executable.
	Name string

	// Args is the base set of arguments to append to when creating a new [Command].
	Args []string

	// Env is a slice of environment variables to set for the created commands.
	// Each environment variable should be in the form "KEY=value".
	Env []string
}

// CreateCommand creates a new [Command] by appending the provided arguments to the
// base command specified in the [ExecAppender]. It also sets the environment
// variables specified in the [ExecAppender] for the created command.
func (a ExecAppender) CreateCommand(ctx context.Context, args ...string) Command {
	args = append(a.Args, args...)
	cmd := exec.CommandContext(ctx, a.Name, args...)
	cmd.Env = append(cmd.Env, a.Env...)
	return cmd
}

var _ Creator = (*ExecAppender)(nil)
