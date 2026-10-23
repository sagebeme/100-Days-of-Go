// Command play runs Lost in Nairobi CBD in your terminal. Write the game loop below,
// then play it from the repo root with:
//
//	go run ./phase-1-terminal/day-014-lost-in-nairobi/starter/cmd/play
package main

import (
	"fmt"

	lostinnairobi "github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-014-lost-in-nairobi/starter"
)

func main() {
	fmt.Println("Umepotea CBD, na mat ya kwenda home iko Kencom. Tafuta njia! (saidia for help, toka to quit)")
	room := lostinnairobi.Start
	fmt.Println(lostinnairobi.Describe(room))

	// Your code here: the game loop. Until the player has won, print "> ", read a line
	// with a bufio.Scanner on os.Stdin, stop on "toka" or the end of input, and pass
	// the line to lostinnairobi.Step. Print the text it returns. Day 12 has the pieces.
}
