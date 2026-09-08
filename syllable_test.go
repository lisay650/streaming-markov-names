package namegen

import (
	"reflect"
	"testing"
)

func TestSplitSyllables(t *testing.T) {
	cases := []struct {
		word string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"ab", []string{"ab"}},
		{"xyz", []string{"xyz"}},
		{"aa", []string{"aa"}},
		{"elena", []string{"e", "le", "na"}},
		{"aria", []string{"a", "ria"}},
		{"bianca", []string{"bian", "ca"}},
		{"brennan", []string{"bren", "nan"}},
	}
	for _, c := range cases {
		if got := splitSyllables(c.word); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitSyllables(%q) = %v, want %v", c.word, got, c.want)
		}
	}
}

func TestSplitSyllables_JoinsBackToOriginalWord(t *testing.T) {
	words := []string{"elena", "aria", "bianca", "brennan", "elliot", "arlo", "x", "bzzt"}
	for _, w := range words {
		got := splitSyllables(w)
		joined := ""
		for _, s := range got {
			joined += s
		}
		if joined != w {
			t.Errorf("splitSyllables(%q) = %v, joined back to %q, want %q", w, got, joined, w)
		}
	}
}
