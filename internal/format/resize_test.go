package format_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"rodusek.dev/pkg/cpp-tools/internal/format"
)

func TestResize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		text    string
		columns int
		want    string
	}{
		{
			name:    "Empty",
			text:    "",
			columns: 20,
			want:    "",
		}, {
			name:    "ZeroColumnsReturnsInputUnchanged",
			text:    "hello world\nsome stuff",
			columns: 0,
			want:    "hello world\nsome stuff",
		}, {
			name:    "NegativeColumnsReturnsInputUnchanged",
			text:    "hello world",
			columns: -5,
			want:    "hello world",
		}, {
			name:    "SingleLineFitsUnchanged",
			text:    "hello world",
			columns: 20,
			want:    "hello world",
		}, {
			name:    "SingleLineWrappedAtWordBoundary",
			text:    "hello world",
			columns: 8,
			want:    "hello\nworld",
		}, {
			name:    "SoftLinebreaksJoinedThenWrapped",
			text:    "hello\nworld",
			columns: 20,
			want:    "hello world",
		}, {
			name:    "ParagraphBreakPreserved",
			text:    "hello\n\nworld",
			columns: 20,
			want:    "hello\n\nworld",
		}, {
			name:    "ThreeNewlinesCollapsedToParagraphBreak",
			text:    "a\n\n\nb",
			columns: 20,
			want:    "a\n\nb",
		}, {
			name:    "LongWordExceedsColumns",
			text:    "verylongwordthatcannotwrap more",
			columns: 8,
			want:    "verylongwordthatcannotwrap\nmore",
		}, {
			name:    "AsteriskBulletSingleItemFits",
			text:    "* hello world",
			columns: 20,
			want:    "* hello world",
		}, {
			name:    "DashBulletSingleItemFits",
			text:    "- hello world",
			columns: 20,
			want:    "- hello world",
		}, {
			name:    "NumberedDotBulletSingleItemFits",
			text:    "1. hello world",
			columns: 20,
			want:    "1. hello world",
		}, {
			name:    "NumberedParenBulletSingleItemFits",
			text:    "1) hello world",
			columns: 20,
			want:    "1) hello world",
		}, {
			name:    "AsteriskBulletContinuationJoinedWhenWide",
			text:    "* hello my\n   baby",
			columns: 20,
			want:    "* hello my baby",
		}, {
			name:    "AsteriskBulletContinuationWrappedWhenNarrow",
			text:    "* hello my\n   baby",
			columns: 8,
			want:    "* hello\n  my\n  baby",
		}, {
			name:    "MultipleAsteriskBullets",
			text:    "* first item\n* second item",
			columns: 20,
			want:    "* first item\n* second item",
		}, {
			name:    "MultipleNumberedBullets",
			text:    "1. first item\n2. second item",
			columns: 20,
			want:    "1. first item\n2. second item",
		}, {
			name:    "NumberedBulletContinuationWrappedWithTwoSpaceIndent",
			text:    "1. hello my baby",
			columns: 8,
			want:    "1. hello\n  my\n  baby",
		}, {
			name:    "DoubleDigitNumberedBullet",
			text:    "10. some text here",
			columns: 20,
			want:    "10. some text here",
		}, {
			name:    "MixedPlainThenBullets",
			text:    "intro paragraph\n\n* first\n* second",
			columns: 20,
			want:    "intro paragraph\n\n* first\n* second",
		}, {
			name:    "AsteriskWithoutSpaceIsNotBullet",
			text:    "*hello",
			columns: 20,
			want:    "*hello",
		}, {
			name:    "DashWithoutSpaceIsNotBullet",
			text:    "-hello",
			columns: 20,
			want:    "-hello",
		}, {
			name:    "NumberWithoutDotOrParenIsNotBullet",
			text:    "1 hello",
			columns: 20,
			want:    "1 hello",
		}, {
			name:    "NumberDotWithoutSpaceIsNotBullet",
			text:    "1.hello",
			columns: 20,
			want:    "1.hello",
		}, {
			name:    "EmptyBulletDroppedToBareMarker",
			text:    "* ",
			columns: 20,
			want:    "*",
		}, {
			name:    "BulletWithVeryLongWordExceedsColumns",
			text:    "* verylongword",
			columns: 6,
			want:    "* verylongword",
		}, {
			name:    "ParagraphOfBulletsThenParagraphOfPlain",
			text:    "* a\n* b\n\nplain text",
			columns: 20,
			want:    "* a\n* b\n\nplain text",
		}, {
			name:    "WrapPacksWordsUntilColumnLimit",
			text:    "one two three four five",
			columns: 10,
			want:    "one two\nthree four\nfive",
		}, {
			name:    "BulletWrapWithExactBoundary",
			text:    "* aa bb cc",
			columns: 5,
			want:    "* aa\n  bb\n  cc",
		}, {
			name:    "WhitespaceOnlyLineBetweenContent",
			text:    "hello\n \nworld",
			columns: 20,
			want:    "hello world",
		}, {
			name:    "BulletColumnsEqualToMarkerClampsFirstWidth",
			text:    "* aa bb",
			columns: 2,
			want:    "* aa\n  bb",
		}, {
			name:    "BulletColumnsBelowContinuationPadClampsRestWidth",
			text:    "* aa bb",
			columns: 1,
			want:    "* aa\n  bb",
		}, {
			name:    "MultipleNumberedBulletsEachWrapped",
			text:    "1. hello my baby\n2. hello my darling",
			columns: 8,
			want:    "1. hello\n  my\n  baby\n2. hello\n  my\n  darling",
		}, {
			name:    "MultipleNumberedBulletsContinuationJoinedThenWrapped",
			text:    "1. hello\n   my baby\n2. hello\n   my darling",
			columns: 20,
			want:    "1. hello my baby\n2. hello my darling",
		}, {
			name:    "MultipleNumberedBulletsContinuationJoinedThenWrappedNarrow",
			text:    "1. hello\n   my baby\n2. hello\n   my darling",
			columns: 10,
			want:    "1. hello\n  my baby\n2. hello\n  my\n  darling",
		}, {
			name:    "MultipleNumberedParenBulletsEachWrapped",
			text:    "1) one two three\n2) four five six",
			columns: 10,
			want:    "1) one two\n  three\n2) four\n  five six",
		}, {
			name:    "MixedDoubleDigitNumberedBulletsWrapped",
			text:    "9. nine items here\n10. ten items here",
			columns: 12,
			want:    "9. nine\n  items here\n10. ten\n  items here",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			text := tc.text
			columns := tc.columns

			// Act
			resized := format.Resize(text, columns)

			// Assert
			if got, want := resized, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Resize(%q, %d) got %q, want %q", text, columns, got, want)
			}
		})
	}
}
