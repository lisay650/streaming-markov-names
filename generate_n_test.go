package namegen

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func TestGenerateN_UntrainedReturnsError(t *testing.T) {
	m, _ := NewModel(2)
	rnd := rand.New(rand.NewSource(1))
	if _, err := m.GenerateN(rnd, 3); !errors.Is(err, ErrUntrained) {
		t.Fatalf("GenerateN on untrained model: err = %v, want %v", err, ErrUntrained)
	}
}

func TestGenerateN_ZeroReturnsNilWithoutError(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(1))
	got, err := m.GenerateN(rnd, 0)
	if err != nil {
		t.Fatalf("GenerateN(rnd, 0): err = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("GenerateN(rnd, 0) = %v, want nil", got)
	}
}

func TestGenerateN_RejectsNegativeCount(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(1))
	if _, err := m.GenerateN(rnd, -1); err == nil {
		t.Fatal("GenerateN(rnd, -1): want error, got nil")
	}
}

func TestGenerateN_ReturnsRequestedCountOfDistinctNames(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(7))
	got, err := m.GenerateN(rnd, 5)
	if err != nil {
		t.Fatalf("GenerateN: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("GenerateN(rnd, 5) returned %d names, want 5", len(got))
	}
	seen := make(map[string]bool)
	for _, name := range got {
		if seen[name] {
			t.Fatalf("GenerateN(rnd, 5) = %v, contains duplicate %q", got, name)
		}
		seen[name] = true
	}
}

// A single-name, order-1 corpus can only ever produce one distinct name, so
// asking for two must exhaust the attempt budget and report the shortfall
// rather than hang or silently return fewer than requested.
func TestGenerateN_ReturnsErrorWhenVocabularyTooSmall(t *testing.T) {
	m, _ := NewModel(1)
	if err := m.Train(strings.NewReader("ab\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(42))
	if _, err := m.GenerateN(rnd, 2); !errors.Is(err, ErrTooFewUniqueNames) {
		t.Fatalf("GenerateN(rnd, 2): err = %v, want %v", err, ErrTooFewUniqueNames)
	}
}

func TestSyllableModel_GenerateNReturnsDistinctNames(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(7))
	got, err := m.GenerateN(rnd, 4)
	if err != nil {
		t.Fatalf("GenerateN: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("GenerateN(rnd, 4) returned %d names, want 4", len(got))
	}
	seen := make(map[string]bool)
	for _, name := range got {
		if seen[name] {
			t.Fatalf("GenerateN(rnd, 4) = %v, contains duplicate %q", got, name)
		}
		seen[name] = true
	}
}
