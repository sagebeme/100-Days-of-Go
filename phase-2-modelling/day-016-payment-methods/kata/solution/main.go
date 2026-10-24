package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type method interface {
	fee(a int) int
	left() int
	take(n int)
}

type wallet struct{ balance int }

func (w *wallet) fee(a int) int {
	switch {
	case a <= 100:
		return 0
	case a <= 500:
		return 7
	case a <= 1000:
		return 13
	case a <= 5000:
		return 33
	default:
		return 57
	}
}
func (w *wallet) left() int  { return w.balance }
func (w *wallet) take(n int) { w.balance -= n }

type card struct{ limit, spent int }

func (c *card) fee(a int) int { return a * 2 / 100 }
func (c *card) left() int     { return c.limit - c.spent }
func (c *card) take(n int)    { c.spent += n }

type cash struct{ notes int }

func (c *cash) fee(a int) int { return 0 }
func (c *cash) left() int     { return c.notes }
func (c *cash) take(n int)    { c.notes -= n }

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	methods := make([]method, n)
	for i := range methods {
		var kind string
		var value int
		fmt.Fscan(in, &kind, &value)
		switch kind {
		case "mpesa":
			methods[i] = &wallet{balance: value}
		case "card":
			methods[i] = &card{limit: value}
		default:
			methods[i] = &cash{notes: value}
		}
	}

	var q int
	fmt.Fscan(in, &q)
	for ; q > 0; q-- {
		var a int
		fmt.Fscan(in, &a)
		best, bestFee := -1, 0
		for i, m := range methods {
			f := m.fee(a)
			if m.left() >= a+f && (best == -1 || f < bestFee) {
				best, bestFee = i, f
			}
		}
		if best == -1 {
			fmt.Fprintln(out, "DECLINED")
			continue
		}
		methods[best].take(a + bestFee)
		fmt.Fprintln(out, best+1, bestFee)
	}

	parts := make([]string, n)
	for i, m := range methods {
		parts[i] = fmt.Sprint(m.left())
	}
	fmt.Fprintln(out, strings.Join(parts, " "))
}
