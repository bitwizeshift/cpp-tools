package spdx_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"rodusek.dev/pkg/cpp-tools/internal/spdx"
)

func TestLookupName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		id     string
		want   string
		wantOK bool
	}{
		{
			name:   "KnownLicense",
			id:     "MIT",
			want:   "MIT License",
			wantOK: true,
		}, {
			name:   "KnownLicenseWithDashes",
			id:     "Apache-2.0",
			want:   "Apache License 2.0",
			wantOK: true,
		}, {
			name:   "KnownException",
			id:     "Classpath-exception-2.0",
			want:   "Classpath exception 2.0",
			wantOK: true,
		}, {
			name:   "UnknownIdentifier",
			id:     "NotARealLicenseID",
			want:   "",
			wantOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			name, ok := spdx.LookupName(tc.id)

			// Assert
			if got, want := name, tc.want; !cmp.Equal(got, want) {
				t.Errorf("LookupName(%q) = name %q, want %q", tc.id, got, want)
			}
			if got, want := ok, tc.wantOK; !cmp.Equal(got, want) {
				t.Errorf("LookupName(%q) = ok %v, want %v", tc.id, got, want)
			}
		})
	}
}

func TestIsKnownIdentifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "KnownLicense",
			id:   "MIT",
			want: true,
		}, {
			name: "KnownException",
			id:   "Classpath-exception-2.0",
			want: true,
		}, {
			name: "Unknown",
			id:   "NotARealLicenseID",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			known := spdx.IsKnownIdentifier(tc.id)

			// Assert
			if got, want := known, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsKnownIdentifier(%q) = %v, want %v", tc.id, got, want)
			}
		})
	}
}

func TestIsKnownLicenseID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "KnownLicense",
			id:   "MIT",
			want: true,
		}, {
			name: "ExceptionIsNotLicense",
			id:   "Classpath-exception-2.0",
			want: false,
		}, {
			name: "Unknown",
			id:   "NotARealLicenseID",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			known := spdx.IsKnownLicenseID(tc.id)

			// Assert
			if got, want := known, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsKnownLicenseID(%q) = %v, want %v", tc.id, got, want)
			}
		})
	}
}

func TestIsKnownExceptionID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "KnownException",
			id:   "Classpath-exception-2.0",
			want: true,
		}, {
			name: "LicenseIsNotException",
			id:   "MIT",
			want: false,
		}, {
			name: "Unknown",
			id:   "NotARealLicenseID",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange

			// Act
			known := spdx.IsKnownExceptionID(tc.id)

			// Assert
			if got, want := known, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsKnownExceptionID(%q) = %v, want %v", tc.id, got, want)
			}
		})
	}
}
