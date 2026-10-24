// Command watch runs Watch against six fake shops started inside this program: one is
// slow, one has a flaky server, one is broken. Nothing goes to the internet.
//
//	go run ./phase-4-concurrency/day-048-price-watcher/starter/cmd/watch
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/sagebeme/100-Days-of-Go/phase-4-concurrency/day-048-price-watcher/starter"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shop := func(body string) string {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, body)
		}))
		return srv.URL // the servers stop when the program ends
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(10 * time.Second):
		case <-r.Context().Done():
		}
	}))
	var calls atomic.Int32
	flaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprint(w, `[{"item":"sukari 1kg","price":175},{"item":"mafuta 1L","price":335}]`)
	}))

	shops := []pricewatcher.Shop{
		{Name: "Mama Njeri Shop", URL: shop(`[{"item":"unga 2kg","price":230},{"item":"sukari 1kg","price":190}]`)},
		{Name: "Duka la Kona", URL: shop(`[{"item":"unga 2kg","price":245},{"item":"sukari 1kg","price":180},{"item":"chumvi 500g","price":30}]`)},
		{Name: "Kariakor Traders", URL: shop(`[{"item":"unga 2kg","price":228},{"item":"mafuta 1L","price":310}]`)},
		{Name: "Gikomba Wholesale", URL: slow.URL},
		{Name: "Ngara Stores", URL: flaky.URL},
		{Name: "Zawadi Mart", URL: shop(`not json`)},
	}
	cfg := pricewatcher.Config{
		Workers: 3,
		Timeout: 500 * time.Millisecond,
		Retries: 3,
		Backoff: 100 * time.Millisecond,
		Limiter: rate.NewLimiter(rate.Every(50*time.Millisecond), 1), // 20 requests a second in all
	}

	start := time.Now()
	summary, err := pricewatcher.Watch(ctx, http.DefaultClient, shops, cfg)
	if err != nil {
		fmt.Println("stopped:", err)
		os.Exit(1)
	}
	fmt.Print(pricewatcher.Report(summary))
	fmt.Printf("(took %s)\n", time.Since(start).Round(10*time.Millisecond))
}
