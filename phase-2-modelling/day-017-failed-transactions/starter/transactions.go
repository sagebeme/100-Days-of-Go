// Package failedtransactions sends money between accounts and explains, in words a
// customer understands, why a transaction failed.
package failedtransactions

import "errors"

// Sentinel errors: compare with errors.Is, never with ==, because Send wraps them.
var (
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrUnknownAccount    = errors.New("unknown account")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrTimeout           = errors.New("network timeout")
)

// LimitError says a payment would take today's total past the daily limit.
type LimitError struct {
	Limit  int // the daily limit, Ksh
	Sent   int // already sent today, Ksh
	Amount int // this payment, Ksh
}

// Error makes *LimitError an error:
//
//	"Ksh3000 would take today's total to Ksh12000, over the Ksh10000 daily limit"
func (e *LimitError) Error() string {
	// Your code here.
	return ""
}

// TxError is the error Send returns. It records which transaction failed and wraps
// the reason.
type TxError struct {
	Code string // the transaction code, e.g. "QFT1ABC2DE"
	Err  error  // why it failed
}

// Error returns "transaction <code>: <reason>".
func (e *TxError) Error() string {
	// Your code here.
	return ""
}

// Unwrap returns the reason, so errors.Is and errors.As can look inside a TxError.
func (e *TxError) Unwrap() error {
	// Your code here.
	return nil
}

// Bank holds every account's balance.
type Bank struct {
	Balances   map[string]int // account → balance in Ksh
	DailyLimit int            // most an account may send in a day, Ksh; 0 means no limit
	SentToday  map[string]int // account → sent so far today, Ksh; Send creates it if nil
}

// Send moves amount shillings from one account to another. Every error it returns is
// a *TxError carrying code, and it checks in this order:
//
//  1. amount below 1: wraps ErrInvalidAmount
//  2. from or to not in Balances: wraps ErrUnknownAccount
//  3. over the daily limit: wraps a *LimitError
//  4. from's balance below amount: wraps ErrInsufficientFunds
//
// On an error nothing changes. On success the money moves and SentToday[from] grows.
func (b *Bank) Send(code, from, to string, amount int) error {
	// Your code here.
	return nil
}

// Explain turns an error into the message a customer sees. If the error contains a
// *TxError, the message starts with its code and a space.
//
//	nil                      "Confirmed."
//	ErrInvalidAmount         "Failed. Enter an amount of at least Ksh1."
//	ErrUnknownAccount        "Failed. Check the number and try again."
//	*LimitError              "Failed. Your daily limit is Ksh10000. You can send Ksh7000 more today."
//	ErrInsufficientFunds     "Failed. You do not have enough money to send this."
//	ErrTimeout               "Failed. Network problem, please try again."
//	anything else            "Failed. Something went wrong."
func Explain(err error) string {
	// Your code here.
	return ""
}

// Retryable reports whether trying the same transaction again might work: only a
// timeout is worth retrying. The rest fail the same way every time.
func Retryable(err error) bool {
	// Your code here.
	return false
}
