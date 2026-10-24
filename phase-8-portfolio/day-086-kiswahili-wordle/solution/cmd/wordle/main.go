// Command wordle plays Kiswahili Wordle in the terminal.
//
//	go run ./phase-8-portfolio/day-086-kiswahili-wordle/solution/cmd/wordle
//	go run ./phase-8-portfolio/day-086-kiswahili-wordle/solution/cmd/wordle -random
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	wordle "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-086-kiswahili-wordle/solution"
)

func main() {
	random := flag.Bool("random", false, "pick a random word instead of today's word")
	flag.Parse()
	color := os.Getenv("NO_COLOR") == ""

	answer := wordle.Daily(time.Now())
	if *random {
		words := wordle.Words()
		answer = words[rand.IntN(len(words))]
	}
	g, err := wordle.NewGame(answer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Karibu! Kisia neno la herufi %d. Una nafasi %d.\n", wordle.Length, wordle.MaxGuesses)
	in := bufio.NewScanner(os.Stdin)
	for !g.Over() {
		fmt.Printf("%d/%d> ", len(g.Turns())+1, wordle.MaxGuesses)
		if !in.Scan() {
			fmt.Println()
			return
		}
		turn, err := g.Guess(in.Text())
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(tiles(turn, color))
		fmt.Println(keyboard(g.Letters(), color))
	}
	if g.Won() {
		fmt.Printf("Hongera! Umelipata kwa jaribio la %d.\n", len(g.Turns()))
	} else {
		fmt.Printf("Pole! Neno lilikuwa %s.\n", strings.ToUpper(g.Answer()))
	}
	fmt.Printf("%s: %s\n", g.Answer(), wordle.Meanings[g.Answer()])
}

// paint wraps s in an ANSI background colour for the mark.
func paint(s string, m wordle.Mark, color bool) string {
	if !color {
		return s
	}
	code := map[wordle.Mark]string{wordle.Correct: "42", wordle.Present: "43", wordle.Absent: "100"}[m]
	return "\x1b[30;" + code + "m" + s + "\x1b[0m"
}

func tiles(t wordle.Turn, color bool) string {
	var b strings.Builder
	for i, m := range t.Marks {
		b.WriteString(paint(" "+strings.ToUpper(t.Guess[i:i+1])+" ", m, color))
	}
	if !color {
		b.WriteString("  " + wordle.Pattern(t.Marks))
	}
	return b.String()
}

func keyboard(seen map[byte]wordle.Mark, color bool) string {
	var b strings.Builder
	for _, row := range []string{"qwertyuiop", "asdfghjkl", "zxcvbnm"} {
		for i := 0; i < len(row); i++ {
			c := row[i]
			m, ok := seen[c]
			switch {
			case !ok:
				b.WriteByte(c)
			case !color && m == wordle.Absent:
				b.WriteByte('_')
			case !color:
				b.WriteByte(c - 'a' + 'A')
			default:
				b.WriteString(paint(string(c), m, color))
			}
			b.WriteByte(' ')
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}
