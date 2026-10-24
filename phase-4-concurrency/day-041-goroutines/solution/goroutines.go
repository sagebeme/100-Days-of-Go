// Package goroutines sends a chama's meeting reminders and adds up its balances, doing
// the slow calls at the same time instead of one after another.
package goroutines

import "sync"

// SendAll calls send once for every phone number, all at the same time, and waits for
// every call to finish. Result i is what send returned for numbers[i], so the results
// come back in the same order as the numbers, whichever send finishes first.
//
//	SendAll([]string{"0712345678", "0723456789"}, send)
//	→ []string{"0712345678: sent", "0723456789: sent"}
func SendAll(numbers []string, send func(number string) string) []string {
	results := make([]string, len(numbers))
	var wg sync.WaitGroup
	for i, n := range numbers {
		wg.Go(func() {
			// Each goroutine writes only its own slot, so no two touch the same memory.
			results[i] = send(n)
		})
	}
	wg.Wait()
	return results
}

// SumBalances asks balance for every account at the same time and returns the total.
//
//	SumBalances([]string{"mpesa", "bank", "sacco"}, balance) → 85500
func SumBalances(accounts []string, balance func(account string) int) int {
	amounts := make([]int, len(accounts))
	var wg sync.WaitGroup
	for i, a := range accounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			amounts[i] = balance(a)
		}()
	}
	wg.Wait()

	// Add up only after Wait: by then every goroutine has written its slot.
	total := 0
	for _, x := range amounts {
		total += x
	}
	return total
}
