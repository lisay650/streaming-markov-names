package namegen

import (
	"bufio"
	"io"
	"strings"
)

// syllableEnd marks the end of a name inside a SyllableModel's transition
// table. It's a string rather than a byte because SyllableModel's tokens are
// whole syllables; "\x00" can't collide with a real syllable produced by
// splitSyllables, since that only ever slices the printable input text.
const syllableEnd = "\x00"

// contextSep joins the syllables that make up a context into a single map
// key. It must not appear inside a syllable itself; splitSyllables only ever
// produces substrings of the trimmed input line, so the unit separator
// control character is safe to use here.
const contextSep = "\x1f"

// SyllableModel is a Markov chain like Model, except its tokens are whole
// syllables (as produced by splitSyllables) rather than single characters.
// The same corpus tends to produce chunkier, more pronounceable output than
// the character-level Model, at the cost of a smaller effective vocabulary
// per context and a requirement that names have at least `order` syllables
// to contribute training data. The zero value is not usable; construct one
// with NewSyllableModel.
type SyllableModel struct {
	order       int
	transitions map[string]map[string]int
	starts      map[string]int
	startTotal  int
}

// NewSyllableModel creates a SyllableModel that predicts each next syllable
// from the preceding `order` syllables. Order 1 is a reasonable default for
// most name corpora, since syllables already carry more shape than single
// characters; higher orders need a larger corpus to have learned enough
// contexts to generalize instead of just replaying training data.
func NewSyllableModel(order int) (*SyllableModel, error) {
	if order < 1 {
		return nil, errOrderTooSmall
	}
	return &SyllableModel{
		order:       order,
		transitions: make(map[string]map[string]int),
		starts:      make(map[string]int),
	}, nil
}

// Train reads names from r, one per line, splits each into syllables, and
// folds the result into the model's transition counts. Like Model.Train, it
// streams the input line by line rather than buffering it all in memory.
func (m *SyllableModel) Train(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}
		m.observe(splitSyllables(strings.ToLower(name)))
	}
	return scanner.Err()
}

func (m *SyllableModel) observe(syllables []string) {
	if len(syllables) < m.order {
		return // too few syllables to seed a context of this order
	}

	start := joinContext(syllables[:m.order])
	m.starts[start]++
	m.startTotal++

	for i := 0; i+m.order < len(syllables); i++ {
		ctx := joinContext(syllables[i : i+m.order])
		next := syllables[i+m.order]
		m.bump(ctx, next)
	}
	lastCtx := joinContext(syllables[len(syllables)-m.order:])
	m.bump(lastCtx, syllableEnd)
}

func (m *SyllableModel) bump(ctx, next string) {
	row, ok := m.transitions[ctx]
	if !ok {
		row = make(map[string]int)
		m.transitions[ctx] = row
	}
	row[next]++
}

// Trained reports whether the model has seen at least one usable name.
func (m *SyllableModel) Trained() bool {
	return m.startTotal > 0
}

func joinContext(syllables []string) string {
	return strings.Join(syllables, contextSep)
}
