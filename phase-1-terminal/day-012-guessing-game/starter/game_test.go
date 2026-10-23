package guessinggame

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestHint(t *testing.T) {
	tests := []struct {
		secret, guess int
		want          string
	}{
		{42, 50, "Too high."},
		{42, 41, "Too low."},
		{42, 42, "Correct!"},
		{1, 100, "Too high."},
		{100, 1, "Too low."},
	}
	for _, tt := range tests {
		got := Hint(tt.secret, tt.guess)
		if got != tt.want {
			t.Errorf("Hint(%d, %d) = %q, want %q", tt.secret, tt.guess, got, tt.want)
		}
	}
}

func TestNewSecretRange(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	seen := map[int]bool{}
	for i := 0; i < 1000; i++ {
		got := NewSecret(r, 10)
		if got < 1 || got > 10 {
			t.Fatalf("NewSecret(r, 10) = %d, want a number from 1 to 10", got)
		}
		seen[got] = true
	}
	if len(seen) != 10 {
		t.Errorf("1000 calls to NewSecret(r, 10) gave only %d different numbers, want all 10", len(seen))
	}
	if got := NewSecret(r, 0); got != 0 {
		t.Errorf("NewSecret(r, 0) = %d, want 0", got)
	}
	if got := NewSecret(r, 1); got != 1 {
		t.Errorf("NewSecret(r, 1) = %d, want 1", got)
	}
}

func TestNewSecretSameSeed(t *testing.T) {
	a := rand.New(rand.NewPCG(7, 7))
	b := rand.New(rand.NewPCG(7, 7))
	for i := 0; i < 20; i++ {
		x, y := NewSecret(a, 100), NewSecret(b, 100)
		if x != y {
			t.Fatalf("call %d: same seed gave %d and %d, want the same number", i+1, x, y)
		}
	}
}

func TestPlay(t *testing.T) {
	tests := []struct {
		about   string
		input   string
		secret  int
		max     int
		tries   int
		wantWon bool
		used    int
		wantOut string
	}{
		{
			"wins on the third guess", "50\n25\n42\n", 42, 100, 7, true, 3,
			"Guess a number from 1 to 100. You have 7 tries.\n" +
				"> Too high.\n" +
				"> Too low.\n" +
				"> Correct! Guesses used: 3.\n",
		},
		{
			"bad lines don't use a try", "abc\n0\n101\n 42 \n", 42, 100, 7, true, 1,
			"Guess a number from 1 to 100. You have 7 tries.\n" +
				"> Numbers only, please.\n" +
				"> From 1 to 100, please.\n" +
				"> From 1 to 100, please.\n" +
				"> Correct! Guesses used: 1.\n",
		},
		{
			"runs out of tries", "1\n2\n3\n", 7, 10, 2, false, 2,
			"Guess a number from 1 to 10. You have 2 tries.\n" +
				"> Too low.\n" +
				"> Too low.\n" +
				"Out of tries. It was 7.\n",
		},
		{
			"input ends early", "9\n", 5, 10, 3, false, 1,
			"Guess a number from 1 to 10. You have 3 tries.\n" +
				"> Too high.\n" +
				"> \nBye! It was 5.\n",
		},
		{
			"no input at all", "", 5, 10, 3, false, 0,
			"Guess a number from 1 to 10. You have 3 tries.\n" +
				"> \nBye! It was 5.\n",
		},
		{
			"last line has no newline", "3\n8", 8, 10, 3, true, 2,
			"Guess a number from 1 to 10. You have 3 tries.\n" +
				"> Too low.\n" +
				"> Correct! Guesses used: 2.\n",
		},
	}
	for _, tt := range tests {
		var out strings.Builder
		won, used := Play(strings.NewReader(tt.input), &out, tt.secret, tt.max, tt.tries)
		if won != tt.wantWon || used != tt.used {
			t.Errorf("%s: Play returned (%v, %d), want (%v, %d)", tt.about, won, used, tt.wantWon, tt.used)
		}
		if out.String() != tt.wantOut {
			t.Errorf("%s: Play printed\n%q\nwant\n%q", tt.about, out.String(), tt.wantOut)
		}
	}
}
