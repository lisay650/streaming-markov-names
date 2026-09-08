package namegen

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func TestSyllableModel_Generate_UntrainedReturnsError(t *testing.T) {
	m, _ := NewSyllableModel(1)
	rnd := rand.New(rand.NewSource(1))
	if _, err := m.Generate(rnd); !errors.Is(err, ErrUntrained) {
		t.Fatalf("Generate on untrained model: err = %v, want %v", err, ErrUntrained)
	}
}

// With a single-name corpus and order 1, every syllable context has exactly
// one possible next syllable, so the walk is forced regardless of the random
// source: it should reproduce the training name, capitalized.
func TestSyllableModel_Generate_DeterministicWithSingleTransitionPerContext(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader("elena\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}

	rnd := rand.New(rand.NewSource(42))
	got, err := m.Generate(rnd)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "Elena" {
		t.Fatalf("Generate() = %q, want %q", got, "Elena")
	}
}

func TestSyllableModel_Generate_CapitalizesFirstLetterOnly(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}

	rnd := rand.New(rand.NewSource(7))
	for i := 0; i < 20; i++ {
		got, err := m.Generate(rnd)
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if got == "" {
			t.Fatal("Generate() returned an empty string")
		}
		if got[:1] != strings.ToUpper(got[:1]) {
			t.Fatalf("Generate() = %q, first letter is not capitalized", got)
		}
		if rest := got[1:]; rest != strings.ToLower(rest) {
			t.Fatalf("Generate() = %q, characters after the first should stay lowercase", got)
		}
	}
}

// A model whose only transition points back to its own context can never
// naturally reach the end token; Generate must still terminate at
// maxSyllables rather than looping forever.
func TestSyllableModel_Generate_StopsAtMaxSyllablesOnCyclicModel(t *testing.T) {
	m, _ := NewSyllableModel(1)
	m.starts["ba"] = 1
	m.startTotal = 1
	m.transitions["ba"] = map[string]int{"ba": 1}

	rnd := rand.New(rand.NewSource(3))
	got, err := m.Generate(rnd)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if want := capitalize(strings.Repeat("ba", maxSyllables)); got != want {
		t.Fatalf("Generate() = %q, want %q", got, want)
	}
}
