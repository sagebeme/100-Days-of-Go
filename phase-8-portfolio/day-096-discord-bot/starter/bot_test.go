package discordbot

import (
	"context"
	"strings"
	"testing"
	"time"
)

func katas() []Kata {
	return []Kata{
		{Day: 95, Name: "Replay Window", Prompt: "ACCEPT or DUPLICATE?", Question: "W=300; evt_a at 100, 250, 400", Answer: "ACCEPT DUPLICATE ACCEPT",
			Hints: []string{"Keep a map from ID to time.", "At exactly W seconds the window is over."}},
		{Day: 93, Name: "Time Travel Store", Prompt: "What was the value?", Question: "SET float 4500 10; GET float 15", Answer: "4500",
			Hints: []string{"Keep every version.", "Binary search the times.", "sort.Search finds the first time after t."}},
		{Day: 94, Name: "Search Suggestions", Prompt: "First three titles.", Question: "devops data developer; prefix de", Answer: "developer devops",
			Hints: []string{"Sort first."}},
	}
}

// on returns a clock stuck at the given date in Nairobi.
func on(year int, month time.Month, day int) func() time.Time {
	eat := time.FixedZone("EAT", 3*60*60)
	return func() time.Time { return time.Date(year, month, day, 8, 0, 0, 0, eat) }
}

func cmd(name, user string, opts ...string) Command {
	c := Command{Name: name, UserID: user, Options: map[string]string{}}
	for i := 0; i+1 < len(opts); i += 2 {
		c.Options[opts[i]] = opts[i+1]
	}
	return c
}

func TestToday(t *testing.T) {
	tests := []struct {
		now  func() time.Time
		want int
	}{
		{on(2026, 1, 1), 93}, // year day 1 -> index 0 in day order
		{on(2026, 1, 2), 94},
		{on(2026, 1, 3), 95},
		{on(2026, 1, 4), 93}, // and round again
		{on(2026, 2, 1), 94}, // year day 32 -> index 31 % 3 = 1
	}
	for _, tt := range tests {
		b := NewBot(katas(), NewMemoryProgress(), tt.now)
		if got := b.Today().Day; got != tt.want {
			t.Errorf("Today() on %s = day %d, want %d", tt.now().Format("2 Jan"), got, tt.want)
		}
	}
	// 23:30 on 1 January in Nairobi is still 1 January, even though it's 20:30 UTC.
	late := func() time.Time { return time.Date(2026, 1, 1, 23, 30, 0, 0, time.FixedZone("EAT", 3*60*60)) }
	if got := NewBot(katas(), NewMemoryProgress(), late).Today().Day; got != 93 {
		t.Errorf("Today() late on 1 Jan in EAT = day %d, want 93", got)
	}
}

func TestKataCommand(t *testing.T) {
	b := NewBot(katas(), NewMemoryProgress(), on(2026, 1, 1))
	tests := []struct {
		cmd           Command
		wantIn        []string
		wantEphemeral bool
	}{
		{cmd("kata", "u1"), []string{"Day 93", "Time Travel Store", "SET float 4500 10; GET float 15", "/submit day:93"}, false},
		{cmd("kata", "u1", "day", "95"), []string{"Day 95", "Replay Window", "ACCEPT or DUPLICATE?"}, false},
		{cmd("kata", "u1", "day", "7"), []string{"No kata for day 7"}, true},
		{cmd("kata", "u1", "day", "abc"), []string{"Usage"}, true},
		{cmd("dance", "u1"), []string{"Unknown command"}, true},
	}
	for _, tt := range tests {
		got := b.Handle(context.Background(), tt.cmd)
		for _, s := range tt.wantIn {
			if !strings.Contains(got.Content, s) {
				t.Errorf("Handle(%+v).Content = %q, want it to contain %q", tt.cmd, got.Content, s)
			}
		}
		if got.Ephemeral != tt.wantEphemeral {
			t.Errorf("Handle(%+v).Ephemeral = %v, want %v", tt.cmd, got.Ephemeral, tt.wantEphemeral)
		}
	}
	if got := b.Handle(context.Background(), cmd("kata", "u1")).Content; strings.Contains(got, "4500\n") || strings.HasSuffix(got, "4500") {
		t.Errorf("/kata must not give away the answer: %q", got)
	}
}

func TestHints(t *testing.T) {
	b := NewBot(katas(), NewMemoryProgress(), on(2026, 1, 1))
	ctx := context.Background()
	steps := []struct {
		cmd  Command
		want string
	}{
		{cmd("hint", "amina", "day", "94"), "Hint 1 of 1: Sort first."},
		{cmd("hint", "amina", "day", "94"), "No more hints"},
		{cmd("hint", "amina", "day", "93"), "Hint 1 of 3: Keep every version."},
		{cmd("hint", "juma", "day", "93"), "Hint 1 of 3: Keep every version."}, // each user has their own count
		{cmd("hint", "amina", "day", "93"), "Hint 2 of 3: Binary search the times."},
		{cmd("hint", "amina", "day", "12"), "No kata for day 12"},
		{cmd("hint", "amina"), "Usage"},
	}
	for i, s := range steps {
		got := b.Handle(ctx, s.cmd)
		if !strings.HasPrefix(got.Content, s.want) {
			t.Errorf("step %d: %s %v by %s = %q, want it to start with %q", i+1, s.cmd.Name, s.cmd.Options, s.cmd.UserID, got.Content, s.want)
		}
		if !got.Ephemeral {
			t.Errorf("step %d: hints must be ephemeral, so others don't see them", i+1)
		}
	}
}

func TestSubmit(t *testing.T) {
	p := NewMemoryProgress()
	b := NewBot(katas(), p, on(2026, 1, 1))
	ctx := context.Background()
	b.Handle(ctx, cmd("hint", "amina", "day", "93"))
	steps := []struct {
		cmd  Command
		want string
	}{
		{cmd("submit", "amina", "day", "93", "answer", "3200"), "Not quite"},
		{cmd("submit", "amina", "day", "93", "answer", "  4500 "), "Correct! Day 93 solved with 1 hints."},
		{cmd("submit", "amina", "day", "93", "answer", "4500"), "Correct! You'd already solved day 93."},
		{cmd("submit", "amina", "day", "95", "answer", "accept duplicate accept"), "Not quite"}, // case matters
		{cmd("submit", "amina", "day", "95", "answer", "ACCEPT   DUPLICATE ACCEPT"), "Correct! Day 95 solved with 0 hints."},
		{cmd("submit", "amina", "day", "93"), "Usage"},
		{cmd("submit", "amina", "day", "93", "answer", "   "), "Usage"},
		{cmd("submit", "amina", "answer", "4500"), "Usage"},
		{cmd("submit", "amina", "day", "50", "answer", "x"), "No kata for day 50"},
	}
	for i, s := range steps {
		got := b.Handle(ctx, s.cmd)
		if !strings.HasPrefix(got.Content, s.want) {
			t.Errorf("step %d: submit %v = %q, want it to start with %q", i+1, s.cmd.Options, got.Content, s.want)
		}
		if !got.Ephemeral {
			t.Errorf("step %d: submit replies must be ephemeral, so answers don't leak", i+1)
		}
	}
	if !p.Solved("amina", 93) || p.Solved("juma", 93) || p.Solved("amina", 94) {
		t.Error("progress should show amina solved 93 and nobody else solved anything else")
	}
}

func TestSameAnswer(t *testing.T) {
	tests := []struct {
		got, want string
		same      bool
	}{
		{"4500", "4500", true},
		{" 4500\t", "4500", true},
		{"a  b", "a b", true},
		{"A b", "a b", false},
		{"4500 3200", "4500", false},
		{"", "4500", false},
	}
	for _, tt := range tests {
		if got := SameAnswer(tt.got, tt.want); got != tt.same {
			t.Errorf("SameAnswer(%q, %q) = %v, want %v", tt.got, tt.want, got, tt.same)
		}
	}
}

// Bot must satisfy Handler, so the HTTP layer can take any Handler.
var _ Handler = (*Bot)(nil)

// Progress is an interface so a real bot can store it in a database.
var _ Progress = (*MemoryProgress)(nil)
