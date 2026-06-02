package spdx_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"rodusek.dev/pkg/cpp-tools/internal/spdx"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		expr    string
		want    *spdx.Expression
		wantErr error
	}{
		{
			name: "SingleLicense",
			expr: "MIT",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "MIT"}},
				}},
			},
		}, {
			name: "LicenseWithOrLater",
			expr: "GPL-2.0+",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "GPL-2.0", OrLater: true}},
				}},
			},
		}, {
			name: "LicenseWithWhitespace",
			expr: "  MIT  ",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "MIT"}},
				}},
			},
		}, {
			name: "AndExpression",
			expr: "MIT AND Apache-2.0",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{
						{ID: "MIT"},
						{ID: "Apache-2.0"},
					},
				}},
			},
		}, {
			name: "OrExpression",
			expr: "MIT OR Apache-2.0",
			want: &spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}}},
					{Licenses: []spdx.License{{ID: "Apache-2.0"}}},
				},
			},
		}, {
			name: "WithExpression",
			expr: "GPL-2.0 WITH Classpath-exception-2.0",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{
						ID:        "GPL-2.0",
						Exception: "Classpath-exception-2.0",
					}},
				}},
			},
		}, {
			name: "WithOnOrLaterLicense",
			expr: "GPL-2.0+ WITH Classpath-exception-2.0",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{
						ID:        "GPL-2.0",
						OrLater:   true,
						Exception: "Classpath-exception-2.0",
					}},
				}},
			},
		}, {
			name: "AndOrPrecedence",
			expr: "MIT AND Apache-2.0 OR BSD-3-Clause",
			want: &spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}, {ID: "Apache-2.0"}}},
					{Licenses: []spdx.License{{ID: "BSD-3-Clause"}}},
				},
			},
		}, {
			name: "ParenthesisedGroupDistributes",
			expr: "(MIT OR Apache-2.0) AND GPL-3.0+",
			want: &spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}, {ID: "GPL-3.0", OrLater: true}}},
					{Licenses: []spdx.License{{ID: "Apache-2.0"}, {ID: "GPL-3.0", OrLater: true}}},
				},
			},
		}, {
			name: "ParenthesisedGroupRight",
			expr: "MIT AND (Apache-2.0 OR BSD-3-Clause)",
			want: &spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}, {ID: "Apache-2.0"}}},
					{Licenses: []spdx.License{{ID: "MIT"}, {ID: "BSD-3-Clause"}}},
				},
			},
		}, {
			name: "NestedParens",
			expr: "((MIT))",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{{ID: "MIT"}},
				}},
			},
		}, {
			name: "AndIsAssociative",
			expr: "MIT AND Apache-2.0 AND BSD-3-Clause",
			want: &spdx.Expression{
				Choices: []spdx.Choice{{
					Licenses: []spdx.License{
						{ID: "MIT"},
						{ID: "Apache-2.0"},
						{ID: "BSD-3-Clause"},
					},
				}},
			},
		}, {
			name: "OrIsAssociative",
			expr: "MIT OR Apache-2.0 OR BSD-3-Clause",
			want: &spdx.Expression{
				Choices: []spdx.Choice{
					{Licenses: []spdx.License{{ID: "MIT"}}},
					{Licenses: []spdx.License{{ID: "Apache-2.0"}}},
					{Licenses: []spdx.License{{ID: "BSD-3-Clause"}}},
				},
			},
		}, {
			name:    "EmptyInput",
			expr:    "",
			wantErr: spdx.ErrInvalidExpression,
		}, {
			name:    "WhitespaceOnly",
			expr:    "   ",
			wantErr: spdx.ErrInvalidExpression,
		}, {
			name:    "InvalidCharacter",
			expr:    "MIT & Apache-2.0",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "DanglingAnd",
			expr:    "MIT AND",
			wantErr: spdx.ErrUnexpectedEOF,
		}, {
			name:    "DanglingOr",
			expr:    "MIT OR",
			wantErr: spdx.ErrUnexpectedEOF,
		}, {
			name:    "UnbalancedOpenParen",
			expr:    "(MIT",
			wantErr: spdx.ErrUnexpectedEOF,
		}, {
			name:    "MissingClosingParen",
			expr:    "(MIT AND Apache-2.0",
			wantErr: spdx.ErrUnexpectedEOF,
		}, {
			name:    "UnbalancedCloseParen",
			expr:    "MIT)",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "WrongClosingToken",
			expr:    "(MIT MIT)",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "UnknownLicense",
			expr:    "FakeLicense-1.0",
			wantErr: spdx.ErrUnknownLicense,
		}, {
			name:    "UnknownException",
			expr:    "GPL-2.0 WITH FakeException-1.0",
			wantErr: spdx.ErrUnknownException,
		}, {
			name:    "WithNonIdentifier",
			expr:    "GPL-2.0 WITH (MIT)",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "WithAtEnd",
			expr:    "GPL-2.0 WITH",
			wantErr: spdx.ErrUnexpectedEOF,
		}, {
			name:    "WithOnNonLicense",
			expr:    "(MIT AND Apache-2.0) WITH Classpath-exception-2.0",
			wantErr: spdx.ErrInvalidExpression,
		}, {
			name:    "WithOnOrExpression",
			expr:    "(MIT OR Apache-2.0) WITH Classpath-exception-2.0",
			wantErr: spdx.ErrInvalidExpression,
		}, {
			name:    "TrailingGarbage",
			expr:    "MIT Apache-2.0",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "LeadingOperator",
			expr:    "AND MIT",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "EmptyParens",
			expr:    "()",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "LeadingPlus",
			expr:    "+",
			wantErr: spdx.ErrUnexpectedToken,
		}, {
			name:    "DoubleWith",
			expr:    "GPL-2.0 WITH Classpath-exception-2.0 WITH Classpath-exception-2.0",
			wantErr: spdx.ErrUnexpectedToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			expression, err := spdx.Parse(tc.expr)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Parse(%q) = error %v, want %v", tc.expr, got, want)
			}
			if got, want := expression, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Parse(%q) = mismatch (-want +got):\n%s", tc.expr, cmp.Diff(want, got))
			}
		})
	}
}
