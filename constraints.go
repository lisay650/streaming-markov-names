package namegen

import (
	"errors"
	"strings"
)

// maxConstrainedAttempts caps how many candidate names GenerateConstrained
// draws before giving up. Rejection sampling is the simplest way to honor
// arbitrary length/prefix/suffix combinations against a transition table
// that has no notion of steering a walk toward a target, but that approach
// needs a backstop for constraints the model's vocabulary can never satisfy.
const maxConstrainedAttempts = 500

// ErrConstraintsNotSatisfiable is returned by GenerateConstrained when no
// candidate drawn from the model matched the given Constraints within
// maxConstrainedAttempts tries.
var ErrConstraintsNotSatisfiable = errors.New("namegen: no generated name satisfied the constraints")

// Constraints narrows the names Generate produces. The zero value places no
// restriction, matching plain Generate.
type Constraints struct {
	// MinLength and MaxLength bound the generated name's length in bytes,
	// inclusive. Zero means unbounded on that side.
	MinLength int
	MaxLength int

	// Prefix and Suffix, if non-empty, must match the start and end of the
	// generated name. The match is case-insensitive, since Generate only
	// ever capitalizes the name's first letter.
	Prefix string
	Suffix string
}

func (c Constraints) validate() error {
	if c.MinLength < 0 || c.MaxLength < 0 {
		return errors.New("namegen: constraint lengths must not be negative")
	}
	if c.MaxLength > 0 && c.MinLength > c.MaxLength {
		return errors.New("namegen: MinLength must not exceed MaxLength")
	}
	return nil
}

func (c Constraints) matches(name string) bool {
	if c.MinLength > 0 && len(name) < c.MinLength {
		return false
	}
	if c.MaxLength > 0 && len(name) > c.MaxLength {
		return false
	}
	if c.Prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(c.Prefix)) {
		return false
	}
	if c.Suffix != "" && !strings.HasSuffix(strings.ToLower(name), strings.ToLower(c.Suffix)) {
		return false
	}
	return true
}

// generateConstrained retries gen, which draws one plain candidate, until
// the result satisfies c or maxConstrainedAttempts is reached. It backs
// Model.GenerateConstrained and SyllableModel.GenerateConstrained, which
// differ only in how a single candidate is drawn.
func generateConstrained(c Constraints, gen func() (string, error)) (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}
	for i := 0; i < maxConstrainedAttempts; i++ {
		name, err := gen()
		if err != nil {
			return "", err
		}
		if c.matches(name) {
			return name, nil
		}
	}
	return "", ErrConstraintsNotSatisfiable
}
