package namegen

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func TestGenerate_UntrainedReturnsError(t *testing.T) {
	m, _ := NewModel(2)
	rnd := rand.New(rand.NewSource(1))
	if _, err := m.Generate(rnd); !errors.Is(err, ErrUntrained) {
		t.Fatalf("Generate on untrained model: err = %v, want %v", err, ErrUntrained)
	}
}

// With a single-name, order-1 corpus every context has exactly one possible
// next byte, so the walk is forced regardless of the random source: it
// should reproduce the training name, capitalized.
func TestGenerate_DeterministicWithSingleTransitionPerContext(t *testing.T) {
	m, _ := NewModel(1)
	if err := m.Train(strings.NewReader("ab\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}

	rnd := rand.New(rand.NewSource(42))
	got, err := m.Generate(rnd)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "Ab" {
		t.Fatalf("Generate() = %q, want %q", got, "Ab")
	}
}

func TestGenerate_CapitalizesFirstLetterOnly(t *testing.T) {
	m, _ := NewModel(2)
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

func TestGenerate_NeverContainsEndSymbolByte(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}

	rnd := rand.New(rand.NewSource(99))
	for i := 0; i < 50; i++ {
		got, err := m.Generate(rnd)
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if strings.IndexByte(got, endSymbol) != -1 {
			t.Fatalf("Generate() = %q, contains raw end-symbol byte", got)
		}
	}
}

// A model whose only transition points back to its own context can never
// naturally reach the end symbol; Generate must still terminate at
// maxNameLength rather than looping forever.
func TestGenerate_StopsAtMaxLengthOnCyclicModel(t *testing.T) {
	m, _ := NewModel(1)
	m.starts["a"] = 1
	m.startTotal = 1
	m.transitions["a"] = map[byte]int{'a': 1}

	rnd := rand.New(rand.NewSource(3))
	got, err := m.Generate(rnd)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(got) != maxNameLength {
		t.Fatalf("Generate() length = %d, want %d", len(got), maxNameLength)
	}
}

func TestCapitalize(t *testing.T) {
	cases := map[string]string{
		"":      "",
		"a":     "A",
		"ab":    "Ab",
		"Ab":    "Ab",
		"abc's": "Abc's",
	}
	for in, want := range cases {
		if got := capitalize(in); got != want {
			t.Errorf("capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}
