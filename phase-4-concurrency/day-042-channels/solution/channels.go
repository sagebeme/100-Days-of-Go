// Package channels turns a stream of M-Pesa till messages into payments and totals,
// passing each value from one stage to the next over channels.
package channels

import (
	"strconv"
	"strings"
)

// Payment is one parsed till message: its transaction code and the amount in shillings.
type Payment struct {
	Code   string
	Amount int
}

// Lines returns a channel that delivers each message in msgs, in order, and is then
// closed. It returns at once: a goroutine does the sending.
//
//	Lines([]string{"QFT1ABC2DE 500", "RKD4XY7Z1P 1250"}) → "QFT1ABC2DE 500", "RKD4XY7Z1P 1250", closed
func Lines(msgs []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out) // the sender closes, once it has nothing more to send
		for _, m := range msgs {
			out <- m
		}
	}()
	return out
}

// Parse reads lines like "QFT1ABC2DE 500" from in and sends a Payment for each one.
// A line that isn't exactly a code and a whole number above 0 is skipped. The returned
// channel is closed after in is closed and every payment has been sent.
//
//	"QFT1ABC2DE 500"  → Payment{"QFT1ABC2DE", 500}
//	"hello"           → skipped
func Parse(in <-chan string) <-chan Payment {
	out := make(chan Payment)
	go func() {
		defer close(out)
		for line := range in {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			amount, err := strconv.Atoi(fields[1])
			if err != nil || amount < 1 {
				continue
			}
			out <- Payment{Code: fields[0], Amount: amount}
		}
	}()
	return out
}

// Total adds up the amounts of every payment from in, until in is closed.
func Total(in <-chan Payment) int {
	total := 0
	for p := range in {
		total += p.Amount
	}
	return total
}

// Batch groups the payments from in into slices of size, in order. The last batch holds
// whatever is left, so it may be shorter. A size below 1 counts as 1. The returned
// channel is closed after the last batch.
//
//	payments A B C D E, size 2 → [A B], [C D], [E], closed
func Batch(in <-chan Payment, size int) <-chan []Payment {
	size = max(size, 1)
	out := make(chan []Payment)
	go func() {
		defer close(out)
		var batch []Payment
		for p := range in {
			batch = append(batch, p)
			if len(batch) == size {
				out <- batch
				batch = nil // a fresh slice, so the receiver keeps the one it got
			}
		}
		if len(batch) > 0 {
			out <- batch
		}
	}()
	return out
}
