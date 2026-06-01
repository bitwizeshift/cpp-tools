package gitdiff

import (
	"context"

	"rodusek.dev/pkg/cpp-tools/internal/command"
)

type ListOptions struct {
	// Indexed indicates whether to list the files in the index instead of the
	// working tree.
	Indexed bool

	// Paths specifies optional path filters to apply to the git diff command.
	// If empty, all changed files are listed. Otherwise, only files matching at
	// the specified paths are listed.
	Paths []string
}

// Lister lists the files changed in a git diff. It uses a [command.Creator] to
// invoke git diff and parses the output. If [Cached] is true, it lists the
// files in the index instead of the working tree.
type Lister struct {
	// CommandCreator is used to create the command that will be executed to list
	// the changed files.
	CommandCreator command.Creator
}

// List lists the files changed in a git diff. It returns a slice of [File]s
// representing the changed files.
func (l *Lister) List(ctx context.Context, opt *ListOptions) ([]File, error) {
	if opt == nil {
		opt = &ListOptions{}
	}
	var args []string
	if opt.Indexed {
		args = append(args, "--cached")
	}
	args = append(args, opt.Paths...)

	cmd := l.CommandCreator.CreateCommand(ctx, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	diff, err := Parse(stdout)
	if err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return diff, nil
}

// DefaultLister is the default [Lister].
var DefaultLister = Lister{
	CommandCreator: command.ExecAppender{
		Name: "git",
		Args: []string{"diff", "--unified=0", "--no-color"},
	},
}
