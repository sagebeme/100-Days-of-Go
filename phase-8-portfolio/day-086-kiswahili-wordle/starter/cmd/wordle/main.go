// Command wordle plays Kiswahili Wordle in the terminal.
//
//	go run ./phase-8-portfolio/day-086-kiswahili-wordle/starter/cmd/wordle
package main

import (
	"fmt"

	wordle "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-086-kiswahili-wordle/starter"
)

func main() {
	// Your code here: pick today's word (or a random one with a -random flag), start a
	// game, then read guesses with bufio.Scanner until g.Over(). Print the tiles and a
	// keyboard after each guess, and the answer with its meaning at the end.
	fmt.Println(wordle.Length)
}
