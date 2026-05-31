package format

import (
	"regexp"
	"strings"
)

var (
	paragraphBreak = regexp.MustCompile(`\n{2,}`)
	numberedBullet = regexp.MustCompile(`^[0-9]+[.)] `)
)

const continuationPad = "  "

// Resize reflows s to fit within columns of width per line, breaking at
// whitespace word boundaries. Words longer than columns are emitted whole and
// exceed the limit. Two or more consecutive newlines are preserved as a
// paragraph break. Lines starting with "* ", "- ", "N." or "N)" (each
// followed by a space) are treated as bullets; wrapped continuation lines of
// a bullet are indented by two spaces.
func Resize(s string, columns int) string {
	if columns <= 0 {
		return s
	}
	paragraphs := paragraphBreak.Split(s, -1)
	out := make([]string, 0, len(paragraphs))
	for _, para := range paragraphs {
		out = append(out, resizeParagraph(para, columns))
	}
	return strings.Join(out, "\n\n")
}

func resizeParagraph(para string, columns int) string {
	blocks := parseBlocks(para)
	rendered := make([]string, 0, len(blocks))
	for _, b := range blocks {
		rendered = append(rendered, renderBlock(b, columns))
	}
	return strings.Join(rendered, "\n")
}

// resizeBlock is a logical run of text within a paragraph: either a single
// bullet item (with its marker and joined continuation text), or a run of
// plain prose.
type resizeBlock struct {
	bullet string
	text   string
}

// parseBlocks groups the lines of a paragraph into resizeBlocks. Bullet lines
// start a new block; subsequent non-bullet lines are absorbed as continuation
// of the current block.
func parseBlocks(para string) []resizeBlock {
	var p paragraphParser
	for line := range strings.SplitSeq(para, "\n") {
		p.consume(line)
	}
	return p.finish()
}

type paragraphParser struct {
	blocks  []resizeBlock
	current resizeBlock
	started bool
}

func (p *paragraphParser) consume(line string) {
	if bullet := detectBullet(line); bullet != "" {
		p.startBullet(bullet, strings.TrimSpace(line[len(bullet):]))
		return
	}
	p.appendContinuation(strings.TrimSpace(line))
}

func (p *paragraphParser) startBullet(bullet, text string) {
	p.flush()
	p.current = resizeBlock{bullet: bullet, text: text}
	p.started = true
}

func (p *paragraphParser) appendContinuation(text string) {
	if !p.started {
		p.current.text = text
		p.started = true
		return
	}
	if text == "" {
		return
	}
	if p.current.text != "" {
		p.current.text += " "
	}
	p.current.text += text
}

func (p *paragraphParser) flush() {
	if !p.started {
		return
	}
	p.blocks = append(p.blocks, p.current)
	p.current = resizeBlock{}
	p.started = false
}

func (p *paragraphParser) finish() []resizeBlock {
	p.flush()
	return p.blocks
}

// detectBullet returns the bullet marker (including the trailing space) if
// line begins with one, or "" if the line is not a bullet.
func detectBullet(line string) string {
	if strings.HasPrefix(line, "* ") {
		return "* "
	}
	if strings.HasPrefix(line, "- ") {
		return "- "
	}
	if m := numberedBullet.FindString(line); m != "" {
		return m
	}
	return ""
}

func renderBlock(b resizeBlock, columns int) string {
	if b.bullet == "" {
		return strings.Join(wrap(b.text, columns, columns), "\n")
	}
	return renderBullet(b, columns)
}

func renderBullet(b resizeBlock, columns int) string {
	firstWidth := atLeast(columns-len(b.bullet), 1)
	restWidth := atLeast(columns-len(continuationPad), 1)
	wrapped := wrap(b.text, firstWidth, restWidth)
	if len(wrapped) == 0 {
		return strings.TrimRight(b.bullet, " ")
	}
	lines := make([]string, 0, len(wrapped))
	lines = append(lines, b.bullet+wrapped[0])
	for _, line := range wrapped[1:] {
		lines = append(lines, continuationPad+line)
	}
	return strings.Join(lines, "\n")
}

// wrap packs the words of text into lines, using firstWidth for the first
// line and restWidth for every subsequent line. A single word longer than the
// applicable width occupies its own line and is allowed to overflow.
func wrap(text string, firstWidth, restWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	current := words[0]
	width := firstWidth
	for _, word := range words[1:] {
		if len(current)+1+len(word) <= width {
			current = current + " " + word
			continue
		}
		lines = append(lines, current)
		current = word
		width = restWidth
	}
	return append(lines, current)
}

func atLeast(v, min int) int {
	if v < min {
		return min
	}
	return v
}
