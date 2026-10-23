// Command play runs the guessing game in your terminal. Once your tests pass, play it
// from the repo root with:
//
//	go run ./phase-1-terminal/day-012-guessing-game/starter/cmd/play
package main

import (
	"math/rand/v2"
	"os"

	guessinggame "github.com/sagebeme/100-Days-of-Go/phase-1-terminal/day-012-guessing-game/starter"
)

func main() {
	const max, tries = 100, 7

	// Seeded from random numbers, so every game has a new secret.
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	secret := guessinggame.NewSecret(r, max)

	guessinggame.Play(os.Stdin, os.Stdout, secret, max, tries)
}
