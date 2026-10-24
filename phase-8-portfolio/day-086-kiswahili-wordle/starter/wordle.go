// Package wordle is a Kiswahili Wordle: guess the five-letter word in six tries.
package wordle

import (
	"errors"
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
	// Your code here
	return nil
}

// IsWord reports whether w is in the list.
func IsWord(w string) bool {
	// Your code here
	return false
}

// Normalize trims spaces and lower-cases a guess.
func Normalize(s string) string {
	// Your code here
	return ""
}

// Score marks guess against answer. Both must be Length bytes of a to z. Greens are
// found first. Then each other letter is yellow only while the answer still has an
// unused copy of it, so a repeated letter is never marked more times than it occurs.
func Score(guess, answer string) [Length]Mark {
	// Your code here
	return [Length]Mark{}
}

// Pattern writes marks as text: G for Correct, Y for Present, . for Absent.
func Pattern(marks [Length]Mark) string {
	// Your code here
	return ""
}

// Daily picks the answer for a date: the same word for everyone all day, a new one
// the next day. Day 0 is 1 January 2026 and the words go round in alphabetical order.
// The date is the one in t's own time zone.
func Daily(t time.Time) string {
	// Your code here
	return ""
}

// Turn is one scored guess.
type Turn struct {
	Guess string
	Marks [Length]Mark
}

// Game is one round.
type Game struct {
	// Your fields here: the answer and the turns so far.
}

// NewGame starts a round. The answer is normalized and must be in the list, or the
// error is ErrUnknown.
func NewGame(answer string) (*Game, error) {
	// Your code here
	return nil, nil
}

// Guess normalizes and scores a guess. A guess of the wrong length, a word not in the
// list, or a guess after the game is over returns an error and uses up no try.
func (g *Game) Guess(s string) (Turn, error) {
	// Your code here
	return Turn{}, nil
}

// Turns returns a copy of the guesses so far.
func (g *Game) Turns() []Turn {
	// Your code here
	return nil
}

// Won reports whether the last guess was the answer.
func (g *Game) Won() bool {
	// Your code here
	return false
}

// Over reports whether the player has won or used every try.
func (g *Game) Over() bool {
	// Your code here
	return false
}

// Answer returns the word. Show it only when the game is over.
func (g *Game) Answer() string {
	// Your code here
	return ""
}

// Letters returns the best mark seen so far for every letter guessed, for drawing a
// keyboard: Correct beats Present beats Absent.
func (g *Game) Letters() map[byte]Mark {
	// Your code here
	return nil
}
