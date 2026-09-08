package namegen

import (
	"errors"
	"strings"
	"testing"
)

func TestNewSyllableModel_RejectsNonPositiveOrder(t *testing.T) {
	for _, order := range []int{0, -1, -5} {
		if _, err := NewSyllableModel(order); err == nil {
			t.Errorf("NewSyllableModel(%d): want error, got nil", order)
		}
	}
}

func TestNewSyllableModel_UntrainedByDefault(t *testing.T) {
	m, err := NewSyllableModel(1)
	if err != nil {
		t.Fatalf("NewSyllableModel: %v", err)
	}
	if m.Trained() {
		t.Fatal("fresh model reports Trained() == true")
	}
}

func TestSyllableModel_Train_MarksModelTrained(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if !m.Trained() {
		t.Fatal("Trained() == false after training on non-empty corpus")
	}
}

func TestSyllableModel_Train_SkipsBlankLines(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader("elena\n\n\naria\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if m.startTotal != 2 {
		t.Fatalf("startTotal = %d, want 2 (blank lines should not count as names)", m.startTotal)
	}
}

func TestSyllableModel_Train_Lowercases(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader("ELENA\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if _, ok := m.starts["e"]; !ok {
		t.Fatalf("starts = %v, want lowercase start context %q", m.starts, "e")
	}
}

func TestSyllableModel_Train_IgnoresNamesWithFewerSyllablesThanOrder(t *testing.T) {
	m, _ := NewSyllableModel(2)
	// "a" and "ab" each split into a single syllable, below order 2.
	if err := m.Train(strings.NewReader("a\nab\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if m.Trained() {
		t.Fatal("Trained() == true after training only on names shorter than the model order")
	}
}

func TestSyllableModel_Train_MultipleCallsBlendCorpora(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader("aria\n")); err != nil {
		t.Fatalf("Train #1: %v", err)
	}
	firstTotal := m.startTotal
	if err := m.Train(strings.NewReader("elena\n")); err != nil {
		t.Fatalf("Train #2: %v", err)
	}
	if m.startTotal != firstTotal+1 {
		t.Fatalf("startTotal = %d after second Train call, want %d", m.startTotal, firstTotal+1)
	}
}

func TestSyllableModel_Train_PropagatesScannerError(t *testing.T) {
	wantErr := errors.New("boom")
	m, _ := NewSyllableModel(1)
	if err := m.Train(errReader{wantErr}); !errors.Is(err, wantErr) {
		t.Fatalf("Train error = %v, want %v", err, wantErr)
	}
}
