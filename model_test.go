package namegen

import (
	"errors"
	"strings"
	"testing"
)

// sampleCorpus is small enough to read at a glance but has enough shared
// prefixes (Ari-, El-) to exercise repeated contexts during Train.
const sampleCorpus = "Aria\nAriel\nArlo\nBianca\nBrennan\nElena\nElliot\n"

func TestNewModel_RejectsNonPositiveOrder(t *testing.T) {
	for _, order := range []int{0, -1, -5} {
		if _, err := NewModel(order); err == nil {
			t.Errorf("NewModel(%d): want error, got nil", order)
		}
	}
}

func TestNewModel_UntrainedByDefault(t *testing.T) {
	m, err := NewModel(2)
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	if m.Trained() {
		t.Fatal("fresh model reports Trained() == true")
	}
}

func TestTrain_MarksModelTrained(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if !m.Trained() {
		t.Fatal("Trained() == false after training on non-empty corpus")
	}
}

func TestTrain_SkipsBlankLines(t *testing.T) {
	m, _ := NewModel(1)
	if err := m.Train(strings.NewReader("ab\n\n\ncd\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if m.startTotal != 2 {
		t.Fatalf("startTotal = %d, want 2 (blank lines should not count as names)", m.startTotal)
	}
}

func TestTrain_TrimsWhitespaceAndCarriageReturn(t *testing.T) {
	m, _ := NewModel(1)
	if err := m.Train(strings.NewReader(" ab \r\ncd\r\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if _, ok := m.starts["a"]; !ok {
		t.Fatalf("starts = %v, want a start context of %q (trailing \\r and spaces should be trimmed)", m.starts, "a")
	}
	if _, ok := m.starts["c"]; !ok {
		t.Fatalf("starts = %v, want a start context of %q", m.starts, "c")
	}
}

func TestTrain_Lowercases(t *testing.T) {
	m, _ := NewModel(1)
	if err := m.Train(strings.NewReader("AB\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if _, ok := m.starts["a"]; !ok {
		t.Fatalf("starts = %v, want lowercase context %q", m.starts, "a")
	}
	if _, ok := m.transitions["a"]['b']; !ok {
		t.Fatalf("transitions[%q] = %v, want lowercase next byte %q", "a", m.transitions["a"], "b")
	}
}

func TestTrain_IgnoresNamesShorterThanOrder(t *testing.T) {
	m, _ := NewModel(3)
	if err := m.Train(strings.NewReader("ab\ncd\n")); err != nil {
		t.Fatalf("Train: %v", err)
	}
	if m.Trained() {
		t.Fatal("Trained() == true after training only on names shorter than the model order")
	}
}

func TestTrain_MultipleCallsBlendCorpora(t *testing.T) {
	m, _ := NewModel(2)
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
	if _, ok := m.starts["ar"]; !ok {
		t.Fatalf("starts = %v, missing context from first Train call", m.starts)
	}
	if _, ok := m.starts["el"]; !ok {
		t.Fatalf("starts = %v, missing context from second Train call", m.starts)
	}
}

type errReader struct{ err error }

func (r errReader) Read([]byte) (int, error) { return 0, r.err }

func TestTrain_PropagatesScannerError(t *testing.T) {
	wantErr := errors.New("boom")
	m, _ := NewModel(2)
	if err := m.Train(errReader{wantErr}); !errors.Is(err, wantErr) {
		t.Fatalf("Train error = %v, want %v", err, wantErr)
	}
}
