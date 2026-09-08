# streaming-markov-names

A Go library for generating random names that sound like a given list of
example names. It works by training a character-level Markov chain on a
corpus of real names (first names, place names, product names, whatever fits
your use case) and then sampling new strings from the learned transitions.

## Why

Most "random name generator" code either ships a small hardcoded word list or
loads an entire corpus file into a slice before doing anything with it. That
second approach falls over once the corpus is large (a full baby-name
dataset, a scraped list of company names, a dictionary) or comes from
somewhere you don't want to buffer in full, like a long-lived network
connection. `Model.Train` reads its input with `bufio.Scanner`, one line at a
time, and folds each line into a small transition-count table as it goes. The
corpus itself is never held in memory, only the (much smaller) model.

## Usage

```go
package main

import (
	"fmt"
	"math/rand"
	"strings"

	namegen "github.com/lisay650/streaming-markov-names"
)

func main() {
	corpus := strings.NewReader("Aria\nAriel\nArlo\nBianca\nBrennan\nElena\nElliot\n")

	model, err := namegen.NewModel(2) // order-2 chain: predict from 2 preceding chars
	if err != nil {
		panic(err)
	}
	if err := model.Train(corpus); err != nil {
		panic(err)
	}

	rnd := rand.New(rand.NewSource(42))
	for i := 0; i < 5; i++ {
		name, err := model.Generate(rnd)
		if err != nil {
			panic(err)
		}
		fmt.Println(name)
	}
}
```

`Train` can be called on any `io.Reader`, including an open `*os.File` for a
multi-gigabyte name list, or the body of an HTTP response:

```go
f, err := os.Open("names.txt")
if err != nil {
	log.Fatal(err)
}
defer f.Close()

model, _ := namegen.NewModel(3)
if err := model.Train(f); err != nil {
	log.Fatal(err)
}
```

You can also call `Train` multiple times, on different readers, to blend
several corpora into one model.

## Choosing an order

The order is how many preceding characters the model looks at to predict the
next one.

- Order 1 mostly ignores the shape of the training names and tends to
  produce noise.
- Order 2 or 3 is a good default: names look invented but plausible.
- Higher orders increasingly just reproduce chunks of the training data
  verbatim.

## Character-level vs. syllable-level

`Model` predicts one character at a time. `SyllableModel` has the same
`Train`/`Generate` shape but predicts one syllable at a time, using a
lightweight heuristic (`splitSyllables`) to break each training name into
syllable-sized chunks before folding them into the transition table:

```go
model, err := namegen.NewSyllableModel(1) // order-1: predict from 1 preceding syllable
if err != nil {
	panic(err)
}
if err := model.Train(corpus); err != nil {
	panic(err)
}

name, err := model.Generate(rnd)
```

Because its tokens are syllables rather than characters, `SyllableModel`
tends to produce output that reads as more pronounceable and less like
random letter noise, at the cost of needing a somewhat larger corpus:
`observe` only learns from a name if it splits into at least `order`
syllables, so very short names contribute nothing at order 2 or above.

## Status

The chain has both a character-level (`Model`) and syllable-level
(`SyllableModel`) implementation. See the roadmap for planned additions like
length/prefix constraints and model serialization.
