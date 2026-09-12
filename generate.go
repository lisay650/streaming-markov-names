package namegen

import (
	"errors"
	"math/rand"
	"strings"
)

// ErrUntrained is returned by Generate when the model has no training data.
var ErrUntrained = errors.New("namegen: model has no training data")

// maxNameLength is a safety valve: with adversarial or degenerate training
// data a walk through the transition table could in principle cycle without
// ever landing on the end symbol.
const maxNameLength = 40

// Generate produces one random name from the model using rnd as the source
// of randomness. Pass a seeded *rand.Rand for reproducible output, or
// rand.New(rand.NewSource(...)) per call site if generating concurrently,
// since *rand.Rand is not safe for concurrent use.
func (m *Model) Generate(rnd *rand.Rand) (string, error) {
	if !m.Trained() {
		return "", ErrUntrained
	}

	ctx := m.pickStart(rnd)
	var sb strings.Builder
	sb.WriteString(ctx)

	for sb.Len() < maxNameLength {
		row := m.transitions[ctx]
		if len(row) == 0 {
			break
		}
		next := pickWeighted(rnd, row)
		if next == endSymbol {
			break
		}
		sb.WriteByte(next)
		ctx = sb.String()[sb.Len()-m.order:]
	}

	return capitalize(sb.String()), nil
}

// GenerateConstrained is like Generate but only returns a name matching c.
// It draws ordinary candidates from Generate and discards the ones that
// don't fit, since the transition table has no way to steer a walk toward a
// target length, prefix, or suffix. If the model's vocabulary can't produce
// a match within a reasonable number of tries, it returns
// ErrConstraintsNotSatisfiable.
func (m *Model) GenerateConstrained(rnd *rand.Rand, c Constraints) (string, error) {
	return generateConstrained(c, func() (string, error) { return m.Generate(rnd) })
}

func (m *Model) pickStart(rnd *rand.Rand) string {
	target := rnd.Intn(m.startTotal)
	for ctx, count := range m.starts {
		target -= count
		if target < 0 {
			return ctx
		}
	}
	panic("namegen: start weights inconsistent with startTotal")
}

func pickWeighted(rnd *rand.Rand, row map[byte]int) byte {
	total := 0
	for _, count := range row {
		total += count
	}
	target := rnd.Intn(total)
	for b, count := range row {
		target -= count
		if target < 0 {
			return b
		}
	}
	panic("namegen: transition weights inconsistent with their total")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
