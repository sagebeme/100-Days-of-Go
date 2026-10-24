// Command server runs the Day 49 server. From the repo root:
//
//	go run ./phase-5-backend/day-049-hello-server/solution/cmd/server
//	PORT=9000 go run ./phase-5-backend/day-049-hello-server/solution/cmd/server
//
// Press Ctrl+C to stop it gracefully.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	helloserver "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-049-hello-server/solution"
)

func main() {
	cfg, err := helloserver.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}

	// ctx is cancelled on Ctrl+C (SIGINT) or on SIGTERM, which is what Docker and
	// most hosts send when they stop a container.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := helloserver.NewServer(cfg.Addr(), helloserver.Routes())

	log.Printf("listening on %s (%s)", ln.Addr(), cfg.AppEnv)
	if err := helloserver.Run(ctx, srv, ln, 10*time.Second); err != nil {
		log.Fatal(err)
	}
	log.Print("stopped cleanly")
}
