// Command play runs Lost in Nairobi CBD in your terminal, using the reference engine.
// Play it from the repo root with:
//
//	go run ./phase-1-terminal/day-014-lost-in-nairobi/solution/cmd/play
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	lostinnairobi "github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-014-lost-in-nairobi/solution"
)

func main() {
	fmt.Println("Umepotea CBD, na mat ya kwenda home iko Kencom. Tafuta njia! (saidia for help, toka to quit)")
	room := lostinnairobi.Start
	fmt.Println(lostinnairobi.Describe(room))

	sc := bufio.NewScanner(os.Stdin)
	for !lostinnairobi.Won(room) {
		fmt.Print("> ")
		if !sc.Scan() {
			fmt.Println()
			return
		}
		command := sc.Text()
		if strings.ToLower(strings.TrimSpace(command)) == "toka" {
			fmt.Println("Sawa, tutaonana.")
			return
		}
		var text string
		room, text = lostinnairobi.Step(room, command)
		fmt.Println(text)
	}
}
