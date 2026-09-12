package namegen

import (
	"math/rand"
	"strings"
)

// maxSyllables is SyllableModel's equivalent of maxNameLength: a safety
// valve against a walk that cycles through contexts without ever landing on
// the end token.
const maxSyllables = 8

// Generate produces one random name from the model using rnd as the source
// of randomness, the same way Model.Generate does but walking a chain of
// syllables instead of characters.
func (m *SyllableModel) Generate(rnd *rand.Rand) (string, error) {
	if !m.Trained() {
		return "", ErrUntrained
	}

	syllables := strings.Split(m.pickStart(rnd), contextSep)
	ctx := syllables

	for len(syllables) < maxSyllables {
		row := m.transitions[joinContext(ctx)]
		if len(row) == 0 {
			break
		}
		next := pickWeightedString(rnd, row)
		if next == syllableEnd {
			break
		}
		syllables = append(syllables, next)
		ctx = syllables[len(syllables)-m.order:]
	}

	return capitalize(strings.Join(syllables, "")), nil
}

// GenerateConstrained is like Generate but only returns a name matching c.
// See Model.GenerateConstrained for how the rejection sampling works and
// when it gives up.
func (m *SyllableModel) GenerateConstrained(rnd *rand.Rand, c Constraints) (string, error) {
	return generateConstrained(c, func() (string, error) { return m.Generate(rnd) })
}

func (m *SyllableModel) pickStart(rnd *rand.Rand) string {
	target := rnd.Intn(m.startTotal)
	for ctx, count := range m.starts {
		target -= count
		if target < 0 {
			return ctx
		}
	}
	panic("namegen: start weights inconsistent with startTotal")
}

func pickWeightedString(rnd *rand.Rand, row map[string]int) string {
	total := 0
	for _, count := range row {
		total += count
	}
	target := rnd.Intn(total)
	for s, count := range row {
		target -= count
		if target < 0 {
			return s
		}
	}
	panic("namegen: transition weights inconsistent with their total")
}
