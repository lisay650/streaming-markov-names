package namegen

import "errors"

// maxGenerateNAttemptsPerName caps how many draws generateN makes per
// requested name before giving up. Like maxConstrainedAttempts, this is a
// backstop for vocabularies too small to produce the requested count of
// distinct names: a one-name corpus can never yield GenerateN(rnd, 2).
const maxGenerateNAttemptsPerName = 50

// ErrTooFewUniqueNames is returned by GenerateN when the model's vocabulary
// can't produce n distinct names within a reasonable number of attempts.
var ErrTooFewUniqueNames = errors.New("namegen: could not generate enough unique names")

// generateN calls gen, which draws one plain candidate, until it has
// collected n names that are pairwise distinct or it exhausts its attempt
// budget. It backs Model.GenerateN and SyllableModel.GenerateN, which differ
// only in how a single candidate is drawn.
func generateN(n int, gen func() (string, error)) ([]string, error) {
	if n < 0 {
		return nil, errors.New("namegen: n must not be negative")
	}
	if n == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, n)
	names := make([]string, 0, n)
	maxAttempts := n * maxGenerateNAttemptsPerName
	for attempts := 0; len(names) < n && attempts < maxAttempts; attempts++ {
		name, err := gen()
		if err != nil {
			return nil, err
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	if len(names) < n {
		return nil, ErrTooFewUniqueNames
	}
	return names, nil
}
