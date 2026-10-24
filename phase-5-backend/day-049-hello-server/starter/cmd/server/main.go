// Command server runs your Day 49 server. Write main below, then run it from the
// repo root with:
//
//	go run ./phase-5-backend/day-049-hello-server/starter/cmd/server
//
// and press Ctrl+C to stop it.
package main

func main() {
	// Your code here. In order:
	//  1. helloserver.LoadConfig(). On an error, print it to os.Stderr and os.Exit(1).
	//  2. ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	//     and defer stop().
	//  3. ln, err := net.Listen("tcp", cfg.Addr())
	//  4. srv := helloserver.NewServer(cfg.Addr(), helloserver.Routes())
	//  5. log the address, then helloserver.Run(ctx, srv, ln, 10*time.Second).
	//
	// Import the package as:
	//  helloserver "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-049-hello-server/starter"
}
