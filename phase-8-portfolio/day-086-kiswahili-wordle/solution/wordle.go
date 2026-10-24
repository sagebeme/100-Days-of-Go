// Package wordle is a Kiswahili Wordle: guess the five-letter word in six tries.
package wordle

import (
	"errors"
	"slices"
	"strings"
	"time"
)

// Length is the number of letters in every word.
const Length = 5

// MaxGuesses is how many tries the player gets.
const MaxGuesses = 6

// Mark is the colour of one tile.
type Mark int

const (
	Absent  Mark = iota // grey: the letter isn't in the word, or not that many times
	Present             // yellow: in the word, somewhere else
	Correct             // green: right letter, right place
)

// Errors that Guess returns. None of them uses up a try.
var (
	ErrWrongLength = errors.New("neno lazima liwe na herufi tano")       // the word must have five letters
	ErrNotAWord    = errors.New("hilo neno halipo kwenye orodha")        // that word isn't in the list
	ErrGameOver    = errors.New("mchezo umekwisha")                      // the game is over
	ErrUnknown     = errors.New("the answer isn't in the list of words") // from NewGame
)

// Words returns every known word in alphabetical order.
func Words() []string {
	words := make([]string, 0, len(Meanings))
	for w := range Meanings {
		words = append(words, w)
	}
	slices.Sort(words)
	return words
}

// IsWord reports whether w is in the list.
func IsWord(w string) bool {
	_, ok := Meanings[w]
	return ok
}

// Normalize trims spaces and lower-cases a guess.
func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Score marks guess against answer. Both must be Length bytes of a to z. Greens are
// found first. Then each other letter is yellow only while the answer still has an
// unused copy of it, so a repeated letter is never marked more times than it occurs.
func Score(guess, answer string) [Length]Mark {
	var marks [Length]Mark
	var left [26]int // answer letters not yet used by a green or a yellow
	for i := range Length {
		if guess[i] == answer[i] {
			marks[i] = Correct
		} else {
			left[answer[i]-'a']++
		}
	}
	for i := range Length {
		if marks[i] == Correct {
			continue
		}
		if c := guess[i] - 'a'; left[c] > 0 {
			marks[i] = Present
			left[c]--
		}
	}
	return marks
}

// Pattern writes marks as text: G for Correct, Y for Present, . for Absent.
func Pattern(marks [Length]Mark) string {
	var b strings.Builder
	for _, m := range marks {
		switch m {
		case Correct:
			b.WriteByte('G')
		case Present:
			b.WriteByte('Y')
		default:
			b.WriteByte('.')
		}
	}
	return b.String()
}

// Daily picks the answer for a date: the same word for everyone all day, a new one
// the next day. Day 0 is 1 January 2026 and the words go round in alphabetical order.
// The date is the one in t's own time zone.
func Daily(t time.Time) string {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	n := int(day.Sub(start).Hours() / 24)
	words := Words()
	i := n % len(words)
	if i < 0 {
		i += len(words)
	}
	return words[i]
}

// Turn is one scored guess.
type Turn struct {
	Guess string
	Marks [Length]Mark
}

// Game is one round.
type Game struct {
	answer string
	turns  []Turn
}

// NewGame starts a round. The answer is normalized and must be in the list, or the
// error is ErrUnknown.
func NewGame(answer string) (*Game, error) {
	answer = Normalize(answer)
	if !IsWord(answer) {
		return nil, ErrUnknown
	}
	return &Game{answer: answer}, nil
}

// Guess normalizes and scores a guess. A guess of the wrong length, a word not in the
// list, or a guess after the game is over returns an error and uses up no try.
func (g *Game) Guess(s string) (Turn, error) {
	if g.Over() {
		return Turn{}, ErrGameOver
	}
	s = Normalize(s)
	if len(s) != Length {
		return Turn{}, ErrWrongLength
	}
	if !IsWord(s) {
		return Turn{}, ErrNotAWord
	}
	t := Turn{Guess: s, Marks: Score(s, g.answer)}
	g.turns = append(g.turns, t)
	return t, nil
}

// Turns returns a copy of the guesses so far.
func (g *Game) Turns() []Turn { return slices.Clone(g.turns) }

// Won reports whether the last guess was the answer.
func (g *Game) Won() bool {
	return len(g.turns) > 0 && g.turns[len(g.turns)-1].Guess == g.answer
}

// Over reports whether the player has won or used every try.
func (g *Game) Over() bool { return g.Won() || len(g.turns) >= MaxGuesses }

// Answer returns the word. Show it only when the game is over.
func (g *Game) Answer() string { return g.answer }

// Letters returns the best mark seen so far for every letter guessed, for drawing a
// keyboard: Correct beats Present beats Absent.
func (g *Game) Letters() map[byte]Mark {
	best := map[byte]Mark{}
	for _, t := range g.turns {
		for i := range Length {
			c := t.Guess[i]
			if m, ok := best[c]; !ok || t.Marks[i] > m {
				best[c] = t.Marks[i]
			}
		}
	}
	return best
}
