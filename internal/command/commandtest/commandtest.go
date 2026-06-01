package commandtest

import (
	"context"
	"errors"
	"io"
	"strings"

	"rodusek.dev/pkg/cpp-tools/internal/command"
)

type state int

const (
	stateNotStarted state = iota
	stateStarted
	stateWaited
)

type fakeCommand struct {
	startErr  error
	waitErr   error
	stdoutErr error
	stderrErr error

	stdout io.ReadCloser
	stderr io.ReadCloser

	state state
}

func (c *fakeCommand) Start() error {
	if c.state != stateNotStarted {
		return errors.New("exec: StdoutPipe after process started")
	}
	c.state = stateStarted
	return c.startErr
}

func (c *fakeCommand) StderrPipe() (io.ReadCloser, error) {
	if c.state >= stateStarted {
		return nil, errors.New("exec: StderrPip after process started")
	}
	return c.stderr, c.stderrErr
}

func (c *fakeCommand) StdoutPipe() (io.ReadCloser, error) {
	if c.state >= stateStarted {
		return nil, errors.New("exec: command already started")
	}
	return c.stdout, c.stdoutErr
}

func (c *fakeCommand) Wait() error {
	if c.state != stateStarted {
		return errors.New("exec: command not started")
	}
	c.state = stateWaited
	return c.waitErr
}

var _ command.Command = (*fakeCommand)(nil)

type commandCreator struct {
	command command.Command
}

func (c commandCreator) CreateCommand(_ context.Context, _ ...string) command.Command {
	return c.command
}

func ErrOnStart(err error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			startErr: err,
		},
	}
}

func ErrOnStderrPipe(err error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			stderrErr: err,
			stdout:    io.NopCloser(strings.NewReader("")),
		},
	}
}

func ErrOnStdoutPipe(err error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			stdoutErr: err,
			stderr:    io.NopCloser(strings.NewReader("")),
		},
	}
}

func ErrOnWait(err error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			waitErr: err,
			stdout:  io.NopCloser(strings.NewReader("")),
			stderr:  io.NopCloser(strings.NewReader("")),
		},
	}
}

func ErrOnPipe(err error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			stderrErr: err,
			stdoutErr: err,
		},
	}
}

func Pipes(stdout, stderr string) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			stdout: io.NopCloser(strings.NewReader(stdout)),
			stderr: io.NopCloser(strings.NewReader(stderr)),
		},
	}
}

func PipesAndWaitErr(stdout, stderr string, waitErr error) command.Creator {
	return commandCreator{
		command: &fakeCommand{
			stdout:  io.NopCloser(strings.NewReader(stdout)),
			stderr:  io.NopCloser(strings.NewReader(stderr)),
			waitErr: waitErr,
		},
	}
}
