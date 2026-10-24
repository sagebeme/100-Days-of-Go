package wordle

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func TestWordList(t *testing.T) {
	words := Words()
	if len(words) != len(Meanings) {
		t.Fatalf("Words() has %d words, want %d", len(words), len(Meanings))
	}
	if !slices.IsSorted(words) {
		t.Errorf("Words() isn't in alphabetical order")
	}
	for _, w := range words {
		if len(w) != Length {
			t.Errorf("%q has %d letters, want %d", w, len(w), Length)
		}
		for i := 0; i < len(w); i++ {
			if w[i] < 'a' || w[i] > 'z' {
				t.Errorf("%q has %q, want only a to z", w, w[i])
			}
		}
	}
	for _, w := range []string{"simba", "nauli", "ugali"} {
		if !IsWord(w) {
			t.Errorf("IsWord(%q) = false, want true", w)
		}
	}
	for _, w := range []string{"", "lion", "abcde", "SIMBA"} {
		if IsWord(w) {
			t.Errorf("IsWord(%q) = true, want false", w)
		}
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		guess, answer, want string
	}{
		{"simba", "simba", "GGGGG"},
		{"mamba", "simba", "..GGG"},
		{"mtoto", "mdomo", "G.G.G"},
		{"ndogo", "mdomo", ".GG.G"},
		{"papai", "punda", "GY..."},
		{"asali", "akili", "G..GG"},
		{"nyama", "mamba", "..YYG"},
		{"ugali", "asali", "..GGG"},
		{"kanga", "ngoma", "..YYG"},
		{"upepo", "pombe", ".YY.Y"},
		{"mkeka", "kesho", ".YY.."},
		{"uhuru", "usiku", "G...G"},
		{"mbuzi", "nzuri", "..GYG"},
		{"tembo", "mbele", ".YYY."},
		{"mamba", "nyama", "YY..G"},
		{"kikoi", "kituo", "GG.Y."},
	}
	for _, tt := range tests {
		if got := Pattern(Score(tt.guess, tt.answer)); got != tt.want {
			t.Errorf("Score(%q, %q) = %s, want %s", tt.guess, tt.answer, got, tt.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	if got := Normalize("  SimBA \n"); got != "simba" {
		t.Errorf("Normalize = %q, want %q", got, "simba")
	}
}

func TestDaily(t *testing.T) {
	nairobi := time.FixedZone("EAT", 3*60*60)
	tests := []struct {
		t    time.Time
		want string
	}{
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "akili"},
		{time.Date(2026, 1, 1, 23, 59, 0, 0, time.UTC), "akili"},
		{time.Date(2026, 1, 2, 8, 0, 0, 0, time.UTC), "amani"},
		{time.Date(2026, 10, 24, 7, 0, 0, 0, nairobi), "tunda"},
		{time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC), "wimbo"},
	}
	for _, tt := range tests {
		if got := Daily(tt.t); got != tt.want {
			t.Errorf("Daily(%v) = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestNewGame(t *testing.T) {
	if _, err := NewGame("lion"); !errors.Is(err, ErrUnknown) {
		t.Errorf("NewGame(lion) error = %v, want ErrUnknown", err)
	}
	g, err := NewGame(" Simba ")
	if err != nil {
		t.Fatalf("NewGame(Simba) error = %v", err)
	}
	if g.Answer() != "simba" {
		t.Errorf("Answer() = %q, want simba", g.Answer())
	}
}

func TestGameWin(t *testing.T) {
	g, _ := NewGame("simba")
	bad := []struct {
		guess string
		want  error
	}{
		{"simb", ErrWrongLength},
		{"simbaa", ErrWrongLength},
		{"abcde", ErrNotAWord},
	}
	for _, b := range bad {
		if _, err := g.Guess(b.guess); !errors.Is(err, b.want) {
			t.Errorf("Guess(%q) error = %v, want %v", b.guess, err, b.want)
		}
	}
	if n := len(g.Turns()); n != 0 {
		t.Fatalf("bad guesses used %d tries, want 0", n)
	}

	turn, err := g.Guess("MAMBA")
	if err != nil || turn.Guess != "mamba" || Pattern(turn.Marks) != "..GGG" {
		t.Fatalf("Guess(MAMBA) = %+v, %v, want mamba ..GGG", turn, err)
	}
	if g.Won() || g.Over() {
		t.Fatalf("after one wrong guess Won/Over = %v/%v, want false/false", g.Won(), g.Over())
	}
	if _, err := g.Guess("simba"); err != nil {
		t.Fatalf("Guess(simba) error = %v", err)
	}
	if !g.Won() || !g.Over() {
		t.Errorf("after the right guess Won/Over = %v/%v, want true/true", g.Won(), g.Over())
	}
	if _, err := g.Guess("tembo"); !errors.Is(err, ErrGameOver) {
		t.Errorf("Guess after winning error = %v, want ErrGameOver", err)
	}
	if n := len(g.Turns()); n != 2 {
		t.Errorf("Turns() has %d, want 2", n)
	}
}

func TestGameLose(t *testing.T) {
	g, _ := NewGame("simba")
	for i, w := range []string{"tembo", "mamba", "punda", "nyoka", "twiga", "nyani"} {
		if g.Over() {
			t.Fatalf("game over after %d guesses, want %d", i, MaxGuesses)
		}
		g.Guess(w)
	}
	if !g.Over() || g.Won() {
		t.Errorf("after 6 wrong guesses Over/Won = %v/%v, want true/false", g.Over(), g.Won())
	}
	if _, err := g.Guess("simba"); !errors.Is(err, ErrGameOver) {
		t.Errorf("a 7th guess error = %v, want ErrGameOver", err)
	}
}

func TestTurnsIsACopy(t *testing.T) {
	g, _ := NewGame("simba")
	g.Guess("tembo")
	g.Turns()[0].Guess = "simba"
	if g.Won() {
		t.Errorf("changing the slice from Turns() changed the game")
	}
}

func TestLetters(t *testing.T) {
	g, _ := NewGame("simba")
	g.Guess("mamba") // ..GGG
	g.Guess("sikia") // GG..G
	want := map[byte]Mark{'m': Correct, 'a': Correct, 'b': Correct, 's': Correct, 'i': Correct, 'k': Absent}
	got := g.Letters()
	if len(got) != len(want) {
		t.Errorf("Letters() = %v, want %v", got, want)
	}
	for c, m := range want {
		if got[c] != m {
			t.Errorf("Letters()[%q] = %v, want %v", c, got[c], m)
		}
	}
}
