// Command proxy runs the rate-limited reverse proxy in front of one upstream.
//
//	go run ./phase-8-portfolio/day-090-rate-limited-proxy/starter/cmd/proxy \
//	    -addr :8080 -prefix /api/ -upstream http://127.0.0.1:9000 -rate 5 -burst 10
package main

import (
	"flag"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"golang.org/x/time/rate"

	proxy "github.com/sagebeme/100-Days-of-Go/phase-8-portfolio/day-090-rate-limited-proxy/starter"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	prefix := flag.String("prefix", "/", "path prefix to forward")
	upstream := flag.String("upstream", "", "upstream base URL (required)")
	perSec := flag.Float64("rate", 5, "requests per second per client")
	burst := flag.Int("burst", 10, "requests a client may make at once")
	trust := flag.String("trust", "", "comma-separated CIDRs of load balancers to trust, e.g. 10.0.0.0/8")
	timeout := flag.Duration("timeout", 10*time.Second, "upstream timeout")
	flag.Parse()

	var trusted []netip.Prefix
	for _, s := range strings.Split(*trust, ",") {
		if s = strings.TrimSpace(s); s != "" {
			trusted = append(trusted, netip.MustParsePrefix(s))
		}
	}
	p, err := proxy.New(proxy.Config{
		Routes:         []proxy.Route{{Prefix: *prefix, Upstream: *upstream}},
		Rate:           rate.Limit(*perSec),
		Burst:          *burst,
		TrustedProxies: trusted,
		Timeout:        *timeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:              *addr,
		Handler:           p,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      *timeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("forwarding %s to %s on %s", *prefix, *upstream, *addr)
	log.Fatal(srv.ListenAndServe())
}
