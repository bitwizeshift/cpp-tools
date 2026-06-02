package spdx_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/spdx"
)

func TestExpression_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		expr spdx.Expression
		want string
	}{
		{
			name: "Empty",
			expr: spdx.Expression{},
			want: "",
		}, {
			name: "SingleLicense",
			expr: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "MIT"}},
				}},
			},
			want: "MIT",
		}, {
			name: "LicenseWithOrLater",
			expr: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "GPL-2.0", OrLater: true}},
				}},
			},
			want: "GPL-2.0+",
		}, {
			name: "AndExpression",
			expr: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{
						{ID: "MIT"},
						{ID: "Apache-2.0"},
					},
				}},
			},
			want: "MIT AND Apache-2.0",
		}, {
			name: "OrExpression",
			expr: spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}}},
					{Licenses: []spdx.License{{ID: "Apache-2.0"}}},
				},
			},
			want: "MIT OR Apache-2.0",
		}, {
			name: "WithException",
			expr: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{
						ID:        "Apache-2.0",
						Exception: "Classpath-exception-2.0",
					}},
				}},
			},
			want: "Apache-2.0 WITH Classpath-exception-2.0",
		}, {
			name: "OrLaterWithException",
			expr: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{
						ID:        "GPL-2.0",
						OrLater:   true,
						Exception: "Classpath-exception-2.0",
					}},
				}},
			},
			want: "GPL-2.0+ WITH Classpath-exception-2.0",
		}, {
			name: "DNFExpanded",
			expr: spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{
						{ID: "MIT"},
						{ID: "GPL-3.0", OrLater: true},
					}},
					{Licenses: []spdx.License{
						{ID: "Apache-2.0"},
						{ID: "GPL-3.0", OrLater: true},
					}},
				},
			},
			want: "MIT AND GPL-3.0+ OR Apache-2.0 AND GPL-3.0+",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			rendered := tc.expr.String()

			// Assert
			if got, want := rendered, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Expression.String() = %q, want %q", got, want)
			}
		})
	}
}

func TestExpression_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		text    string
		want    spdx.Expression
		wantErr error
	}{
		{
			name: "SingleLicense",
			text: "MIT",
			want: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "MIT"}},
				}},
			},
			wantErr: nil,
		}, {
			name: "AndExpression",
			text: "MIT AND Apache-2.0",
			want: spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{
						{ID: "MIT"},
						{ID: "Apache-2.0"},
					},
				}},
			},
			wantErr: nil,
		}, {
			name:    "EmptyInput",
			text:    "",
			want:    spdx.Expression{},
			wantErr: spdx.ErrInvalidExpression,
		}, {
			name:    "UnknownLicense",
			text:    "NotARealLicenseID",
			want:    spdx.Expression{},
			wantErr: spdx.ErrUnknownLicense,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var expr spdx.Expression

			// Act
			err := expr.UnmarshalText([]byte(tc.text))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Expression.UnmarshalText(%q) error = %v, want %v", tc.text, got, want)
			}
			if got, want := expr, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Expression.UnmarshalText(%q) = mismatch (-want +got):\n%s", tc.text, cmp.Diff(want, got))
			}
		})
	}
}
