// Package mpesasms builds the confirmation text M-Pesa sends after a payment.
package mpesasms

// BuildConfirmation returns the text for a payment of amount shillings to name.
//
//	BuildConfirmation("QFT1ABC2DE", 500, "Amina Wanjiru")
//	→ "QFT1ABC2DE Confirmed. Ksh500.00 sent to Amina Wanjiru."
func BuildConfirmation(code string, amount int, name string) string {
	// Your code here. Follow the steps in the README.
	return ""
}
