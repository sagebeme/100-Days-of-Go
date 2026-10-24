package channels

import (
	"reflect"
	"slices"
	"testing"
	"time"
)

// guard is how long a test waits for something that should happen at once. Correct code
// never gets near it; it only stops a stuck test from hanging for ten minutes.
const guard = 2 * time.Second

// start calls f in a goroutine and returns its channel, failing the test if f doesn't
// return quickly or returns a nil channel.
func start[T any](t *testing.T, name string, f func() <-chan T) <-chan T {
	t.Helper()
	got := make(chan (<-chan T), 1)
	go func() { got <- f() }()
	select {
	case ch := <-got:
		if ch == nil {
			t.Fatalf("%s returned a nil channel. Make one with make(chan ...)", name)
		}
		return ch
	case <-time.After(guard):
		t.Fatalf("%s didn't return. Send from a goroutine, so the function can return the channel straight away", name)
		return nil
	}
}

// collect receives everything from ch until it is closed.
func collect[T any](t *testing.T, name string, ch <-chan T) []T {
	t.Helper()
	var all []T
	timeout := time.After(guard)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return all
			}
			all = append(all, v)
		case <-timeout:
			t.Fatalf("%s's channel was never closed (received %v so far). Close it when the sending is done", name, all)
			return nil
		}
	}
}

// feed returns a channel that delivers vals and is then closed.
func feed[T any](vals ...T) <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for _, v := range vals {
			ch <- v
		}
	}()
	return ch
}

func TestLines(t *testing.T) {
	tests := []struct {
		name string
		msgs []string
	}{
		{"two messages", []string{"QFT1ABC2DE 500", "RKD4XY7Z1P 1250"}},
		{"one message", []string{"SBC9QWE1RT 20"}},
		{"no messages", nil},
		{"keeps the order", []string{"a", "b", "c", "d", "e", "f"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := start(t, "Lines", func() <-chan string { return Lines(tt.msgs) })
			got := collect(t, "Lines", ch)
			if !slices.Equal(got, tt.msgs) {
				t.Errorf("Lines(%q) delivered %q, want %q", tt.msgs, got, tt.msgs)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []Payment
	}{
		{"good lines", []string{"QFT1ABC2DE 500", "RKD4XY7Z1P 1250"},
			[]Payment{{"QFT1ABC2DE", 500}, {"RKD4XY7Z1P", 1250}}},
		{"extra spaces are fine", []string{"  QFT1ABC2DE    500  "}, []Payment{{"QFT1ABC2DE", 500}}},
		{"skips bad lines", []string{
			"hello",
			"QFT1ABC2DE 500",
			"RKD4XY7Z1P mia tano",
			"SBC9QWE1RT 0",
			"TUV2ASD3FG -40",
			"",
			"ABC1DEF2GH 300 extra",
			"WXY3ZAB4CD 75",
		}, []Payment{{"QFT1ABC2DE", 500}, {"WXY3ZAB4CD", 75}}},
		{"nothing in", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := start(t, "Parse", func() <-chan Payment { return Parse(feed(tt.lines...)) })
			got := collect(t, "Parse", ch)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q)\n got: %v\nwant: %v", tt.lines, got, tt.want)
			}
		})
	}
}

func TestTotal(t *testing.T) {
	tests := []struct {
		name     string
		payments []Payment
		want     int
	}{
		{"three payments", []Payment{{"A", 500}, {"B", 1250}, {"C", 20}}, 1770},
		{"one payment", []Payment{{"A", 99}}, 99},
		{"none", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan int, 1)
			go func() { done <- Total(feed(tt.payments...)) }()
			select {
			case got := <-done:
				if got != tt.want {
					t.Errorf("Total(%v) = %d, want %d", tt.payments, got, tt.want)
				}
			case <-time.After(guard):
				t.Fatal("Total never returned. Range over the channel: the loop ends when it is closed")
			}
		})
	}
}

func TestBatch(t *testing.T) {
	a, b, c, d, e := Payment{"A", 1}, Payment{"B", 2}, Payment{"C", 3}, Payment{"D", 4}, Payment{"E", 5}
	tests := []struct {
		name     string
		payments []Payment
		size     int
		want     [][]Payment
	}{
		{"last batch is short", []Payment{a, b, c, d, e}, 2, [][]Payment{{a, b}, {c, d}, {e}}},
		{"batches fit exactly", []Payment{a, b, c, d}, 2, [][]Payment{{a, b}, {c, d}}},
		{"one big batch", []Payment{a, b, c}, 10, [][]Payment{{a, b, c}}},
		{"size 1", []Payment{a, b}, 1, [][]Payment{{a}, {b}}},
		{"size 0 counts as 1", []Payment{a, b}, 0, [][]Payment{{a}, {b}}},
		{"nothing in, no batches", nil, 3, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := start(t, "Batch", func() <-chan []Payment { return Batch(feed(tt.payments...), tt.size) })
			got := collect(t, "Batch", ch)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Batch(%v, %d)\n got: %v\nwant: %v", tt.payments, tt.size, got, tt.want)
			}
		})
	}
}

func TestPipeline(t *testing.T) {
	msgs := []string{"QFT1ABC2DE 500", "spam: you have won!", "RKD4XY7Z1P 1250", "SBC9QWE1RT 20"}
	done := make(chan int, 1)
	go func() { done <- Total(Parse(Lines(msgs))) }()
	select {
	case got := <-done:
		if got != 1770 {
			t.Errorf("Total(Parse(Lines(%q))) = %d, want 1770", msgs, got)
		}
	case <-time.After(guard):
		t.Fatal("the pipeline never finished. Check that every stage closes its channel")
	}
}
