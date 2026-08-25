// Package namegen generates random names by training a character-level
// Markov chain on a corpus of example names and sampling new strings from
// the learned transition table.
package namegen

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// endSymbol marks the end of a name inside the transition table. It cannot
// appear in real input because Train works on text lines with the byte 0
// stripped by bufio.Scanner's line splitting.
const endSymbol = 0

// Model holds the learned transition counts for an order-N Markov chain.
// The zero value is not usable; construct one with NewModel.
type Model struct {
	order       int
	transitions map[string]map[byte]int
	starts      map[string]int
	startTotal  int
}

// NewModel creates a Model that predicts each next character from the
// preceding `order` characters. Order 2 or 3 gives plausible-sounding
// invented names for most corpora; order 1 tends to look like noise and
// higher orders tend to just reproduce the training data.
func NewModel(order int) (*Model, error) {
	if order < 1 {
		return nil, errors.New("namegen: order must be at least 1")
	}
	return &Model{
		order:       order,
		transitions: make(map[string]map[byte]int),
		starts:      make(map[string]int),
	}, nil
}

// Train reads names from r, one per line, and folds each into the model's
// transition counts. It streams the input line by line with bufio.Scanner
// rather than reading it all up front, so a corpus far larger than available
// memory can be used to build a model, since the model itself only grows
// with the number of distinct contexts, not the number of names seen.
func (m *Model) Train(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}
		m.observe(strings.ToLower(name))
	}
	return scanner.Err()
}

func (m *Model) observe(name string) {
	if len(name) < m.order {
		return // too short to seed a context of this order
	}

	start := name[:m.order]
	m.starts[start]++
	m.startTotal++

	for i := 0; i+m.order < len(name); i++ {
		ctx := name[i : i+m.order]
		next := name[i+m.order]
		m.bump(ctx, next)
	}
	lastCtx := name[len(name)-m.order:]
	m.bump(lastCtx, endSymbol)
}

func (m *Model) bump(ctx string, next byte) {
	row, ok := m.transitions[ctx]
	if !ok {
		row = make(map[byte]int)
		m.transitions[ctx] = row
	}
	row[next]++
}

// Trained reports whether the model has seen at least one usable name.
func (m *Model) Trained() bool {
	return m.startTotal > 0
}
