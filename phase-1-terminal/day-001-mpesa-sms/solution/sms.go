// Package mpesasms builds the confirmation text M-Pesa sends after a payment.
package mpesasms

import "fmt"

// BuildConfirmation returns the text for a payment of amount shillings to name.
//
//	BuildConfirmation("QFT1ABC2DE", 500, "Amina Wanjiru")
//	→ "QFT1ABC2DE Confirmed. Ksh500.00 sent to Amina Wanjiru."
func BuildConfirmation(code string, amount int, name string) string {
	const status = "Confirmed."
	const currency = "Ksh"

	message := code + " " + status
	message = message + fmt.Sprintf(" %s%d.00 sent to %s.", currency, amount, name)
	return message
}
