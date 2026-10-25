// Package discordbot is a bot that serves the course katas in a Discord server. The
// command logic (Bot) knows nothing about Discord: it takes a Command and returns a
// Response. A thin HTTP layer (interactions.go) turns Discord's interaction requests
// into Commands, so every rule here is tested without Discord.
package discordbot

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Command is one slash command a user ran, e.g. /submit day:93 answer:4500.
type Command struct {
	Name    string            // "kata", "hint" or "submit"
	Options map[string]string // option name to value, as text
	UserID  string
}

// Response is what the bot says back. Ephemeral replies are seen only by the user who
// ran the command.
type Response struct {
	Content   string
	Ephemeral bool
}

// Handler answers commands. Bot is one; tests and other front ends can add their own.
type Handler interface {
	Handle(ctx context.Context, c Command) Response
}

// Kata is one exercise the bot can serve. Answer is a one-line answer to Question,
// because a Discord string option holds one line.
type Kata struct {
	Day      int
	Name     string
	Prompt   string
	Question string
	Answer   string
	Hints    []string
}

// Progress stores what each user has done. MemoryProgress is the in-memory one; a real
// bot would keep this in Postgres or Redis.
type Progress interface {
	HintsUsed(user string, day int) int
	SetHintsUsed(user string, day, n int)
	Solved(user string, day int) bool
	MarkSolved(user string, day int)
}

// Bot is the command logic.
type Bot struct {
	katas    []Kata // sorted by Day
	progress Progress
	now      func() time.Time
}

// NewBot returns a bot serving katas (in any order), storing progress in p and using
// now to pick today's kata.
func NewBot(katas []Kata, p Progress, now func() time.Time) *Bot {
	k := slices.Clone(katas)
	slices.SortFunc(k, func(a, b Kata) int { return cmp.Compare(a.Day, b.Day) })
	return &Bot{katas: k, progress: p, now: now}
}

// Today returns the kata for now's date: kata number (YearDay-1) mod len(katas), in
// day order, so it changes at midnight and cycles through them all. Use the date in
// now's own location.
func (b *Bot) Today() Kata {
	if len(b.katas) == 0 {
		return Kata{}
	}
	return b.katas[(b.now().YearDay()-1)%len(b.katas)]
}

func (b *Bot) find(day int) (Kata, bool) {
	i, ok := slices.BinarySearchFunc(b.katas, day, func(k Kata, d int) int { return cmp.Compare(k.Day, d) })
	if !ok {
		return Kata{}, false
	}
	return b.katas[i], true
}

// Handle runs one command:
//
//	/kata [day]           the kata (today's without day), visible to everyone
//	/hint day             the user's next hint: "Hint 1 of 3: ...", then "No more hints"
//	/submit day answer    "Correct! ..." or "Not quite. ..."; a solved kata stays solved
//
// /hint and /submit reply ephemerally. A missing or bad day gives a usage message, an
// unknown one "No kata for day N", and an unknown command "Unknown command".
func (b *Bot) Handle(ctx context.Context, c Command) Response {
	switch c.Name {
	case "kata":
		k := b.Today()
		if v, ok := c.Options["day"]; ok {
			var resp Response
			if k, resp, ok = b.kataOption(v); !ok {
				return resp
			}
		}
		if k.Day == 0 {
			return private("No katas loaded yet.")
		}
		return Response{Content: fmt.Sprintf("**Day %d: %s**\n%s\nQuestion: `%s`\nAnswer with `/submit day:%d answer:...`",
			k.Day, k.Name, k.Prompt, k.Question, k.Day)}
	case "hint":
		k, resp, ok := b.kataOption(c.Options["day"])
		if !ok {
			return resp
		}
		used := b.progress.HintsUsed(c.UserID, k.Day)
		if used >= len(k.Hints) {
			return private(fmt.Sprintf("No more hints for day %d. You've seen all %d.", k.Day, len(k.Hints)))
		}
		b.progress.SetHintsUsed(c.UserID, k.Day, used+1)
		return private(fmt.Sprintf("Hint %d of %d: %s", used+1, len(k.Hints), k.Hints[used]))
	case "submit":
		k, resp, ok := b.kataOption(c.Options["day"])
		if !ok {
			return resp
		}
		answer, ok := c.Options["answer"]
		if !ok || strings.TrimSpace(answer) == "" {
			return private("Usage: /submit day:<number> answer:<your answer>")
		}
		if !SameAnswer(answer, k.Answer) {
			return private(fmt.Sprintf("Not quite. Try `/hint day:%d`.", k.Day))
		}
		if b.progress.Solved(c.UserID, k.Day) {
			return private(fmt.Sprintf("Correct! You'd already solved day %d.", k.Day))
		}
		b.progress.MarkSolved(c.UserID, k.Day)
		return private(fmt.Sprintf("Correct! Day %d solved with %d hints.", k.Day, b.progress.HintsUsed(c.UserID, k.Day)))
	}
	return private("Unknown command. Try /kata, /hint or /submit.")
}

// kataOption parses a day option and finds its kata. When ok is false, resp says why.
func (b *Bot) kataOption(v string) (k Kata, resp Response, ok bool) {
	day, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || day < 1 {
		return Kata{}, private("Usage: give a day number, e.g. day:93"), false
	}
	k, found := b.find(day)
	if !found {
		return Kata{}, private(fmt.Sprintf("No kata for day %d.", day)), false
	}
	return k, Response{}, true
}

func private(s string) Response { return Response{Content: s, Ephemeral: true} }

// SameAnswer compares two answers the way the kata judge does for one line: it ignores
// spaces at the ends and how many spaces sit between words, but not case.
func SameAnswer(got, want string) bool {
	return slices.Equal(strings.Fields(got), strings.Fields(want))
}

// MemoryProgress keeps progress in memory. It is safe for concurrent use.
type MemoryProgress struct {
	mu     sync.Mutex
	hints  map[string]int
	solved map[string]bool
}

// NewMemoryProgress returns empty progress.
func NewMemoryProgress() *MemoryProgress {
	return &MemoryProgress{hints: map[string]int{}, solved: map[string]bool{}}
}

func key(user string, day int) string { return user + "/" + strconv.Itoa(day) }

// HintsUsed returns how many hints user has seen for day.
func (m *MemoryProgress) HintsUsed(user string, day int) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hints[key(user, day)]
}

// SetHintsUsed records that user has seen n hints for day.
func (m *MemoryProgress) SetHintsUsed(user string, day, n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hints[key(user, day)] = n
}

// Solved reports whether user has solved day.
func (m *MemoryProgress) Solved(user string, day int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.solved[key(user, day)]
}

// MarkSolved records that user solved day.
func (m *MemoryProgress) MarkSolved(user string, day int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.solved[key(user, day)] = true
}
