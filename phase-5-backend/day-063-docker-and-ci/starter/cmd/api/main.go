// Command api runs the Twende health endpoints the way the real API will run in a
// container. Follow step 6 in the README.
//
//	go run ./phase-5-backend/day-063-docker-and-ci/starter/cmd/api
package main

import (
	"fmt"
	"os"

	dockerandci "github.com/sagebeme/100-Days-of-Go/phase-5-backend/day-063-docker-and-ci/starter"
)

func main() {
	cfg, err := dockerandci.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config:\n%v\n", err)
		os.Exit(1)
	}
	// Your code here:
	// 1. a -healthcheck flag that GETs http://127.0.0.1:<port>/healthz and exits 0 or 1
	// 2. a Health with a "db" check when cfg.DatabaseURL is set
	// 3. an http.Server with timeouts, started in a goroutine
	// 4. signal.NotifyContext for SIGINT and SIGTERM; on a signal, SetDraining,
	//    then srv.Shutdown with cfg.ShutdownTimeout
	fmt.Println("config ok, port", cfg.Port)
}
