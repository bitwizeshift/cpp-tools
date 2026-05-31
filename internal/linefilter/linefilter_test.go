package linefilter_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"rodusek.dev/pkg/cpp-tools/internal/linefilter"
)

func TestLineFilter_Includes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		filter linefilter.LineFilters
		file   string
		line   int
		want   bool
	}{
		{
			name:   "EmptyFilter",
			filter: linefilter.LineFilters{},
			file:   "foo.cpp",
			line:   1,
			want:   false,
		}, {
			name: "MatchingFileAndLine",
			filter: linefilter.LineFilters{
				{
					Name: "foo.cpp",
					Lines: []linefilter.Range{
						{Start: 1, End: 1},
					},
				},
			},
			file: "foo.cpp",
			line: 1,
			want: true,
		}, {
			name: "MatchingFileNotLine",
			filter: linefilter.LineFilters{
				{
					Name: "foo.cpp",
					Lines: []linefilter.Range{
						{Start: 2, End: 3},
					},
				},
			},
			file: "foo.cpp",
			line: 1,
			want: false,
		}, {
			name: "FilterOnlyHasName",
			filter: linefilter.LineFilters{
				{
					Name: "foo.cpp",
				},
			},
			file: "foo.cpp",
			line: 1,
			want: true,
		}, {
			name: "FilterNameDoesNotMatch",
			filter: linefilter.LineFilters{
				{
					Name: "bar.cpp",
					Lines: []linefilter.Range{
						{Start: 1, End: 1},
					},
				},
			},
			file: "foo.cpp",
			line: 1,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.filter

			// Act
			includes := sut.Includes(tc.file, tc.line)

			// Assert
			if got, want := includes, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Includes(%q, %d) got %v, want %v", tc.file, tc.line, got, want)
			}
		})
	}
}

func TestRange_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		json    string
		want    linefilter.Range
		wantErr error
	}{
		{
			name:    "IncorrectJSON",
			json:    `{"start": 1, "end": 2}`,
			wantErr: linefilter.ErrFormat,
		},
		{
			name: "CorrectRange",
			json: `[1,2]`,
			want: linefilter.Range{Start: 1, End: 2},
		}, {
			name: "StartEqualsEnd",
			json: `[3,3]`,
			want: linefilter.Range{Start: 3, End: 3},
		}, {
			name:    "TooFewElements",
			json:    `[1]`,
			wantErr: linefilter.ErrFormat,
		}, {
			name:    "TooManyElements",
			json:    `[1,2,3]`,
			wantErr: linefilter.ErrFormat,
		}, {
			name:    "NegativeStart",
			json:    `[-1,2]`,
			wantErr: linefilter.ErrDomain,
		}, {
			name:    "NegativeEnd",
			json:    `[1,-2]`,
			wantErr: linefilter.ErrDomain,
		}, {
			name:    "StartGreaterThanEnd",
			json:    `[5,4]`,
			wantErr: linefilter.ErrDomain,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var sut linefilter.Range

			// Act
			err := json.Unmarshal([]byte(tc.json), &sut)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("UnmarshalJSON(%q) got error %v, want %v", tc.json, got, want)
			}
		})
	}
}
