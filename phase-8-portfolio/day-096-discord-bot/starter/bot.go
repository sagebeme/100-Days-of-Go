// Package discordbot is a bot that serves the course katas in a Discord server. The
// command logic (Bot) knows nothing about Discord: it takes a Command and returns a
// Response. A thin HTTP layer (interactions.go) turns Discord's interaction requests
// into Commands, so every rule here is tested without Discord.
package discordbot

import (
	"context"
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

// Bot is the command logic. Add the fields you need.
type Bot struct {
}

// NewBot returns a bot serving katas (in any order), storing progress in p and using
// now to pick today's kata.
func NewBot(katas []Kata, p Progress, now func() time.Time) *Bot {
	// Your code here
	return &Bot{}
}

// Today returns the kata for now's date: kata number (YearDay-1) mod len(katas), in
// day order, so it changes at midnight and cycles through them all. Use the date in
// now's own location.
func (b *Bot) Today() Kata {
	// Your code here
	return Kata{}
}

// Handle runs one command:
//
//	/kata [day]         the kata (today's without day), visible to everyone:
//	                    "**Day 93: Time Travel Store**", the prompt, the question,
//	                    and "/submit day:93 answer:..." (never the answer)
//	/hint day           the user's next hint: "Hint 1 of 3: ...", then "No more hints ..."
//	/submit day answer  "Not quite. ...", or "Correct! Day 93 solved with 1 hints.",
//	                    or "Correct! You'd already solved day 93."
//
// /hint and /submit reply ephemerally. A missing or bad day, or a missing or blank
// answer, gives a message starting "Usage"; an unknown day "No kata for day N."; an
// unknown command "Unknown command. ..." (ephemeral too).
func (b *Bot) Handle(ctx context.Context, c Command) Response {
	// Your code here
	return Response{}
}

// SameAnswer compares two answers the way the kata judge does for one line: it ignores
// spaces at the ends and how many spaces sit between words, but not case.
func SameAnswer(got, want string) bool {
	// Your code here
	return false
}

// MemoryProgress keeps progress in memory. It is safe for concurrent use.
type MemoryProgress struct {
}

// NewMemoryProgress returns empty progress.
func NewMemoryProgress() *MemoryProgress {
	// Your code here
	return &MemoryProgress{}
}

// HintsUsed returns how many hints user has seen for day.
func (m *MemoryProgress) HintsUsed(user string, day int) int {
	// Your code here
	return 0
}

// SetHintsUsed records that user has seen n hints for day.
func (m *MemoryProgress) SetHintsUsed(user string, day, n int) {
	// Your code here
}

// Solved reports whether user has solved day.
func (m *MemoryProgress) Solved(user string, day int) bool {
	// Your code here
	return false
}

// MarkSolved records that user solved day.
func (m *MemoryProgress) MarkSolved(user string, day int) {
	// Your code here
}
