package diagnostic_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"rodusek.dev/pkg/cpp-tools/internal/diagnostic"
)

func TestNewGitHubLogger_ReturnsNonNilLogger(t *testing.T) {
	t.Parallel()

	// Arrange
	w := &bytes.Buffer{}

	// Act
	logger := diagnostic.NewGitHubLogger(w)

	// Assert
	if got, want := logger == nil, false; !cmp.Equal(got, want) {
		t.Errorf("NewGitHubLogger(...) got %v, want %v", got, want)
	}
}

func TestGitHubHandler_Enabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{
			name:  "Debug",
			level: slog.LevelDebug,
			want:  true,
		}, {
			name:  "Info",
			level: slog.LevelInfo,
			want:  true,
		}, {
			name:  "Warn",
			level: slog.LevelWarn,
			want:  true,
		}, {
			name:  "Error",
			level: slog.LevelError,
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			handler := diagnostic.NewGitHubHandler(&bytes.Buffer{})

			// Act
			enabled := handler.Enabled(context.Background(), tc.level)

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GitHubHandler.Enabled(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestGitHubHandler_WithAttrs_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	handler := diagnostic.NewGitHubHandler(&bytes.Buffer{})

	// Act
	newHandler := handler.WithAttrs([]slog.Attr{slog.String("k", "v")})

	// Assert
	if got, want := newHandler == slog.Handler(handler), true; !cmp.Equal(got, want) {
		t.Errorf("GitHubHandler.WithAttrs(...) got %v, want %v", got, want)
	}
}

func TestGitHubHandler_WithGroup_ReturnsReceiver(t *testing.T) {
	t.Parallel()

	// Arrange
	handler := diagnostic.NewGitHubHandler(&bytes.Buffer{})

	// Act
	newHandler := handler.WithGroup("g")

	// Assert
	if got, want := newHandler == slog.Handler(handler), true; !cmp.Equal(got, want) {
		t.Errorf("GitHubHandler.WithGroup(...) got %v, want %v", got, want)
	}
}

func TestGitHubHandler_Handle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		level     slog.Level
		id        string
		title     string
		message   string
		file      string
		lineStart int64
		lineEnd   int64
		colStart  int64
		colEnd    int64
		want      string
	}{
		{
			name:      "ErrorMessageOnly",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::boom\n",
		}, {
			name:      "WarnLevelEmitsWarningSeverity",
			level:     slog.LevelWarn,
			id:        "",
			title:     "",
			message:   "careful",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::warning::careful\n",
		}, {
			name:      "DebugLevelEmitsDebugSeverity",
			level:     slog.LevelDebug,
			id:        "",
			title:     "",
			message:   "tracing",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::debug::tracing\n",
		}, {
			name:      "InfoLevelEmitsInfoSeverity",
			level:     slog.LevelInfo,
			id:        "",
			title:     "",
			message:   "fyi",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::info::fyi\n",
		}, {
			name:      "UnknownLevelDefaultsToInfo",
			level:     slog.LevelInfo + 1,
			id:        "",
			title:     "",
			message:   "fyi",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::info::fyi\n",
		}, {
			name:      "TitleWithoutIDEmitsBareTitle",
			level:     slog.LevelError,
			id:        "",
			title:     "broken",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=broken::boom\n",
		}, {
			name:      "TitleWithIDEmitsBracketedID",
			level:     slog.LevelError,
			id:        "E001",
			title:     "broken",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=[E001] broken::boom\n",
		}, {
			name:      "IDWithoutTitleOmitsTitleProperty",
			level:     slog.LevelError,
			id:        "E001",
			title:     "",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::boom\n",
		}, {
			name:      "FileEmitsFileProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "main.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error file=main.cpp::boom\n",
		}, {
			name:      "AllPropertiesCommaSeparated",
			level:     slog.LevelError,
			id:        "E001",
			title:     "broken",
			message:   "boom",
			file:      "main.cpp",
			lineStart: 10,
			lineEnd:   12,
			colStart:  5,
			colEnd:    8,
			want:      "::error title=[E001] broken,file=main.cpp,line=10,endLine=12,col=5,endColumn=8::boom\n",
		}, {
			name:      "LineStartOnly",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "a.cpp",
			lineStart: 10,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error file=a.cpp,line=10::boom\n",
		}, {
			name:      "LineEndOnly",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "a.cpp",
			lineStart: 0,
			lineEnd:   12,
			colStart:  0,
			colEnd:    0,
			want:      "::error file=a.cpp,endLine=12::boom\n",
		}, {
			name:      "ColStartOnly",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "a.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  5,
			colEnd:    0,
			want:      "::error file=a.cpp,col=5::boom\n",
		}, {
			name:      "ColEndOnly",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "a.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    8,
			want:      "::error file=a.cpp,endColumn=8::boom\n",
		}, {
			name:      "MessageNewlineEscapedAsPercent0A",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "line1\nline2",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::line1%0Aline2\n",
		}, {
			name:      "MessageCarriageReturnEscapedAsPercent0D",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "line1\rline2",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::line1%0Dline2\n",
		}, {
			name:      "MessagePercentEscapedAsPercent25",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "100% done",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::100%25 done\n",
		}, {
			name:      "MessageColonNotEscaped",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "a:b",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::a:b\n",
		}, {
			name:      "MessageCommaNotEscaped",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "a,b",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error::a,b\n",
		}, {
			name:      "TitleColonEscapedInProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "a:b",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=a%3Ab::boom\n",
		}, {
			name:      "TitleCommaEscapedInProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "a,b",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=a%2Cb::boom\n",
		}, {
			name:      "TitlePercentEscapedInProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "50%",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=50%25::boom\n",
		}, {
			name:      "TitleNewlineEscapedInProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "a\nb",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=a%0Ab::boom\n",
		}, {
			name:      "TitleCarriageReturnEscapedInProperty",
			level:     slog.LevelError,
			id:        "",
			title:     "a\rb",
			message:   "boom",
			file:      "",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error title=a%0Db::boom\n",
		}, {
			name:      "FileWithCommaEscaped",
			level:     slog.LevelError,
			id:        "",
			title:     "",
			message:   "boom",
			file:      "weird,name.cpp",
			lineStart: 0,
			lineEnd:   0,
			colStart:  0,
			colEnd:    0,
			want:      "::error file=weird%2Cname.cpp::boom\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			handler := diagnostic.NewGitHubHandler(&buf)
			rec := diagnosticRecord(
				tc.level, tc.id, tc.title, tc.message,
				tc.file, tc.lineStart, tc.lineEnd, tc.colStart, tc.colEnd,
			)

			// Act
			err := handler.Handle(context.Background(), rec)
			output := buf.String()

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("GitHubHandler.Handle(...) got err=%v, want %v", got, want)
			}
			if got, want := output, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GitHubHandler.Handle(...) got %q, want %q", got, want)
			}
		})
	}
}

type githubFailingWriter struct{ err error }

func (f githubFailingWriter) Write([]byte) (int, error) { return 0, f.err }

func TestGitHubHandler_Handle_WriterError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")
	testCases := []struct {
		name    string
		writeIn error
		want    error
	}{
		{
			name:    "WriterReturnsSentinelError",
			writeIn: testErr,
			want:    testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			handler := diagnostic.NewGitHubHandler(githubFailingWriter{err: tc.writeIn})
			rec := diagnosticRecord(slog.LevelError, "", "", "boom", "", 0, 0, 0, 0)

			// Act
			err := handler.Handle(context.Background(), rec)

			// Assert
			if got, want := err, tc.want; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("GitHubHandler.Handle(...) got err=%v, want %v", got, want)
			}
		})
	}
}
