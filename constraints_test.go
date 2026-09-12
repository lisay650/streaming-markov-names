package namegen

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
)

func TestConstraints_ValidateRejectsNegativeLengths(t *testing.T) {
	for _, c := range []Constraints{
		{MinLength: -1},
		{MaxLength: -1},
	} {
		if err := c.validate(); err == nil {
			t.Errorf("Constraints{%+v}.validate(): want error, got nil", c)
		}
	}
}

func TestConstraints_ValidateRejectsMinAboveMax(t *testing.T) {
	c := Constraints{MinLength: 5, MaxLength: 3}
	if err := c.validate(); err == nil {
		t.Fatal("validate(): want error when MinLength > MaxLength, got nil")
	}
}

func TestConstraints_Matches(t *testing.T) {
	cases := []struct {
		name string
		c    Constraints
		want bool
	}{
		{"Aria", Constraints{}, true},
		{"Aria", Constraints{MinLength: 5}, false},
		{"Aria", Constraints{MinLength: 4}, true},
		{"Aria", Constraints{MaxLength: 3}, false},
		{"Aria", Constraints{MaxLength: 4}, true},
		{"Aria", Constraints{Prefix: "ar"}, true},
		{"Aria", Constraints{Prefix: "AR"}, true},
		{"Aria", Constraints{Prefix: "el"}, false},
		{"Aria", Constraints{Suffix: "ia"}, true},
		{"Aria", Constraints{Suffix: "IA"}, true},
		{"Aria", Constraints{Suffix: "on"}, false},
		{"Aria", Constraints{Prefix: "ar", Suffix: "ia", MinLength: 4, MaxLength: 4}, true},
	}
	for _, tc := range cases {
		if got := tc.c.matches(tc.name); got != tc.want {
			t.Errorf("Constraints{%+v}.matches(%q) = %v, want %v", tc.c, tc.name, got, tc.want)
		}
	}
}

func TestGenerate_ConstrainedRejectsInvalidConstraints(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(1))
	if _, err := m.GenerateConstrained(rnd, Constraints{MinLength: 5, MaxLength: 3}); err == nil {
		t.Fatal("GenerateConstrained with MinLength > MaxLength: want error, got nil")
	}
}

func TestGenerate_ConstrainedHonorsPrefixSuffixAndLength(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(7))
	c := Constraints{Prefix: "ar", MinLength: 3, MaxLength: 10}
	for i := 0; i < 10; i++ {
		got, err := m.GenerateConstrained(rnd, c)
		if err != nil {
			t.Fatalf("GenerateConstrained: %v", err)
		}
		if !c.matches(got) {
			t.Fatalf("GenerateConstrained() = %q, does not satisfy %+v", got, c)
		}
	}
}

func TestGenerate_ConstrainedGivesUpOnUnsatisfiableConstraints(t *testing.T) {
	m, _ := NewModel(2)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(7))
	c := Constraints{Prefix: "zzzzz"}
	if _, err := m.GenerateConstrained(rnd, c); !errors.Is(err, ErrConstraintsNotSatisfiable) {
		t.Fatalf("GenerateConstrained() err = %v, want %v", err, ErrConstraintsNotSatisfiable)
	}
}

func TestSyllableModel_GenerateConstrainedHonorsSuffix(t *testing.T) {
	m, _ := NewSyllableModel(1)
	if err := m.Train(strings.NewReader(sampleCorpus)); err != nil {
		t.Fatalf("Train: %v", err)
	}
	rnd := rand.New(rand.NewSource(7))
	c := Constraints{Suffix: "a"}
	for i := 0; i < 10; i++ {
		got, err := m.GenerateConstrained(rnd, c)
		if err != nil {
			t.Fatalf("GenerateConstrained: %v", err)
		}
		if !c.matches(got) {
			t.Fatalf("GenerateConstrained() = %q, does not satisfy %+v", got, c)
		}
	}
}
