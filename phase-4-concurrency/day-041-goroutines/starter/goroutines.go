// Package goroutines sends a chama's meeting reminders and adds up its balances, doing
// the slow calls at the same time instead of one after another.
package goroutines

// SendAll calls send once for every phone number, all at the same time, and waits for
// every call to finish. Result i is what send returned for numbers[i], so the results
// come back in the same order as the numbers, whichever send finishes first.
//
//	SendAll([]string{"0712345678", "0723456789"}, send)
//	→ []string{"0712345678: sent", "0723456789: sent"}
func SendAll(numbers []string, send func(number string) string) []string {
	// Your code here. Follow the steps in the README.
	return nil
}

// SumBalances asks balance for every account at the same time and returns the total.
//
//	SumBalances([]string{"mpesa", "bank", "sacco"}, balance) → 85500
func SumBalances(accounts []string, balance func(account string) int) int {
	// Your code here.
	return 0
}
