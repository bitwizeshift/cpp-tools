package modified

import (
	"bufio"
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"rodusek.dev/pkg/cpp-tools/internal/command"
)

// YearRanger reports the inclusive range of years during which a path was
// modified. Implementations typically derive this from version-control history.
type YearRanger interface {
	// YearRange returns the earliest and latest modification years for path.
	// If no modification history is available, start and end may be zero.
	YearRange(ctx context.Context, path string) (start, end int, err error)
}

// CommandYearRanger is a [YearRanger] that derives modification years by
// running an external command. The command is expected to write one year per
// line to stdout; blank and non-numeric lines are ignored.
type CommandYearRanger struct {
	// CommandCreator constructs the command that emits modification years for
	// the queried path. The path is passed to the command as its sole argument.
	CommandCreator command.Creator

	// Now returns the current time. It is used as a fallback source of the
	// current year when the command produces no usable output.
	Now func() time.Time
}

// YearRange implements [YearRanger] by invoking the configured command and
// parsing one year per line from its stdout. It returns the minimum and
// maximum year observed, or an error if the command could not be started.
func (cr CommandYearRanger) YearRange(ctx context.Context, path string) (start, end int, err error) {
	command := cr.CommandCreator.CreateCommand(ctx, path)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return 0, 0, err
	}
	if err := command.Start(); err != nil {
		return 0, 0, err
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		year, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		if year < start || start == 0 {
			start = year
		}
		if year > end {
			end = year
		}
	}
	if err := command.Wait(); err != nil {
		return 0, 0, err
	}
	if end == 0 && start == 0 {
		return 0, 0, errors.New("no modification years found")
	}
	return
}

var _ YearRanger = (*CommandYearRanger)(nil)

// GitYearRanger is a [CommandYearRanger] preconfigured to derive modification
// years from `git log` for a given path.
var GitYearRanger = CommandYearRanger{
	CommandCreator: command.ExecAppender{
		Name: "git",
		Args: []string{"log", "--follow", "--format=%ad", "--date=format:%Y", "--"},
	},
	Now: time.Now,
}
