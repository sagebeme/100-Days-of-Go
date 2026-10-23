// Package guessinggame is a number guessing game that reads guesses from any input and
// writes replies to any output, so it can be played in a terminal or checked by a test.
package guessinggame

import (
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
	"strings"
)

// Hint compares guess with secret and returns "Too low.", "Too high." or "Correct!".
//
//	Hint(42, 50) → "Too high."
func Hint(secret, guess int) string {
	switch {
	case guess < secret:
		return "Too low."
	case guess > secret:
		return "Too high."
	default:
		return "Correct!"
	}
}

// NewSecret picks a whole number from 1 to max using r. It returns 0 if max is below 1.
// Two generators made with the same seed give the same secrets, which is what makes
// the game testable.
//
//	r := rand.New(rand.NewPCG(1, 2))
//	NewSecret(r, 100) → a number from 1 to 100
func NewSecret(r *rand.Rand, max int) int {
	if max < 1 {
		return 0
	}
	return r.IntN(max) + 1
}

// Play runs one game. It reads one guess per line from in and writes every message to
// out. A line that isn't a number, or is outside 1 to max, gets a reminder and doesn't
// use up a try. Play stops on the right answer, after tries guesses, or when in runs
// out of lines. It reports whether the player won and how many tries they used.
func Play(in io.Reader, out io.Writer, secret, max, tries int) (won bool, used int) {
	sc := bufio.NewScanner(in)
	fmt.Fprintf(out, "Guess a number from 1 to %d. You have %d tries.\n", max, tries)
	for used < tries {
		fmt.Fprint(out, "> ")
		if !sc.Scan() {
			fmt.Fprintf(out, "\nBye! It was %d.\n", secret)
			return false, used
		}
		guess, err := strconv.Atoi(strings.TrimSpace(sc.Text()))
		if err != nil {
			fmt.Fprintln(out, "Numbers only, please.")
			continue
		}
		if guess < 1 || guess > max {
			fmt.Fprintf(out, "From 1 to %d, please.\n", max)
			continue
		}
		used++
		if guess == secret {
			fmt.Fprintf(out, "Correct! Guesses used: %d.\n", used)
			return true, used
		}
		fmt.Fprintln(out, Hint(secret, guess))
	}
	fmt.Fprintf(out, "Out of tries. It was %d.\n", secret)
	return false, used
}
