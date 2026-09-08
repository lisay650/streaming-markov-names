package namegen

// isVowel reports whether b is one of the five English vowel letters. It
// intentionally does not treat 'y' as a vowel: for short invented names 'y'
// much more often acts like a consonant (as in "yara", "bryn"), and folding
// it in tends to merge syllables that should stay separate.
func isVowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

// splitSyllables breaks a lowercase word into syllable-sized chunks using a
// standard maximal-vowel-run heuristic: find each run of consecutive vowels,
// then place a boundary in the consonant run between two vowel runs. A lone
// consonant between vowels goes with the syllable that follows it (open
// syllable), and a run of two or more consonants splits after the first one.
// This is not a dictionary-accurate syllabifier, just consistent enough to
// give SyllableModel tokens shorter than a whole word and longer than a
// single character.
//
// The input is assumed to already be lowercase, matching how Model.Train and
// SyllableModel.Train normalize their corpus before this is called.
func splitSyllables(word string) []string {
	if word == "" {
		return nil
	}

	type vowelRun struct{ start, end int }
	var runs []vowelRun
	inRun := false
	for i := 0; i < len(word); i++ {
		if isVowel(word[i]) {
			if !inRun {
				runs = append(runs, vowelRun{start: i})
				inRun = true
			}
		} else if inRun {
			runs[len(runs)-1].end = i
			inRun = false
		}
	}
	if inRun {
		runs[len(runs)-1].end = len(word)
	}

	if len(runs) < 2 {
		return []string{word}
	}

	boundaries := make([]int, 0, len(runs)-1)
	for i := 0; i < len(runs)-1; i++ {
		gapStart, gapEnd := runs[i].end, runs[i+1].start
		if gapEnd-gapStart >= 2 {
			boundaries = append(boundaries, gapStart+1)
		} else {
			boundaries = append(boundaries, gapStart)
		}
	}

	syllables := make([]string, 0, len(boundaries)+1)
	prev := 0
	for _, b := range boundaries {
		syllables = append(syllables, word[prev:b])
		prev = b
	}
	return append(syllables, word[prev:])
}
