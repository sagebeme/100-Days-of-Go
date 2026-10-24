// Command sms sends a few fake reminders two per second, so you can watch Throttle pace
// them with a real time.Ticker. Press Ctrl+C to stop early.
//
//	go run ./phase-4-concurrency/day-047-rate-limiting/solution/cmd/sms
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sagebeme/100-Days-of-Go/phase-4-concurrency/day-047-rate-limiting/solution"
)

func main() {
	// The context is cancelled when you press Ctrl+C.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	numbers := []string{"0712000001", "0712000002", "0712000003", "0712000004", "0712000005", "0712000006"}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	err := ratelimiting.Throttle(ctx, numbers, ticker.C, func(n string) {
		fmt.Printf("%5s  reminder sent to %s\n", time.Since(start).Round(100*time.Millisecond), n)
	})
	if err != nil {
		fmt.Println("stopped:", err)
		os.Exit(1)
	}
}
