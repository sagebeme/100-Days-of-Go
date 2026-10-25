// Command api runs the matatu booking API.
//
//	JWT_SECRET=$(openssl rand -base64 32) ADMIN_PHONES=0712345678 \
//	  go run ./phase-5-backend/day-064-matatu-booking-api/starter/cmd/api
package main

import (
	"fmt"
	"os"

	api "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-064-matatu-booking-api/starter"
)

func main() {
	cfg, err := api.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	// Your code here: a MemStore with SeedRoutes, NewServer with a FakePayer, an
	// http.Server with timeouts on cfg.Port, and a graceful shutdown on SIGINT or
	// SIGTERM that calls SetDraining before srv.Shutdown (Day 63).
	_ = cfg
}
