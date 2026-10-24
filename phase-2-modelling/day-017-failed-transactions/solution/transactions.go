// Package failedtransactions sends money between accounts and explains, in words a
// customer understands, why a transaction failed.
package failedtransactions

import (
	"errors"
	"fmt"
)

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
	return fmt.Sprintf("Ksh%d would take today's total to Ksh%d, over the Ksh%d daily limit",
		e.Amount, e.Sent+e.Amount, e.Limit)
}

// TxError is the error Send returns. It records which transaction failed and wraps
// the reason.
type TxError struct {
	Code string // the transaction code, e.g. "QFT1ABC2DE"
	Err  error  // why it failed
}

// Error returns "transaction <code>: <reason>".
func (e *TxError) Error() string {
	return "transaction " + e.Code + ": " + e.Err.Error()
}

// Unwrap returns the reason, so errors.Is and errors.As can look inside a TxError.
func (e *TxError) Unwrap() error {
	return e.Err
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
	if err := b.check(from, to, amount); err != nil {
		return &TxError{Code: code, Err: err}
	}
	if b.SentToday == nil {
		b.SentToday = map[string]int{}
	}
	b.Balances[from] -= amount
	b.Balances[to] += amount
	b.SentToday[from] += amount
	return nil
}

func (b *Bank) check(from, to string, amount int) error {
	if amount < 1 {
		return fmt.Errorf("%w: Ksh%d", ErrInvalidAmount, amount)
	}
	for _, acc := range []string{from, to} {
		if _, ok := b.Balances[acc]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknownAccount, acc)
		}
	}
	sent := b.SentToday[from]
	if b.DailyLimit > 0 && sent+amount > b.DailyLimit {
		return &LimitError{Limit: b.DailyLimit, Sent: sent, Amount: amount}
	}
	if bal := b.Balances[from]; bal < amount {
		return fmt.Errorf("%w: balance Ksh%d, need Ksh%d", ErrInsufficientFunds, bal, amount)
	}
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
	if err == nil {
		return "Confirmed."
	}
	prefix := ""
	var tx *TxError
	if errors.As(err, &tx) {
		prefix = tx.Code + " "
	}
	var limit *LimitError
	switch {
	case errors.Is(err, ErrInvalidAmount):
		return prefix + "Failed. Enter an amount of at least Ksh1."
	case errors.Is(err, ErrUnknownAccount):
		return prefix + "Failed. Check the number and try again."
	case errors.As(err, &limit):
		return fmt.Sprintf("%sFailed. Your daily limit is Ksh%d. You can send Ksh%d more today.",
			prefix, limit.Limit, limit.Limit-limit.Sent)
	case errors.Is(err, ErrInsufficientFunds):
		return prefix + "Failed. You do not have enough money to send this."
	case errors.Is(err, ErrTimeout):
		return prefix + "Failed. Network problem, please try again."
	default:
		return prefix + "Failed. Something went wrong."
	}
}

// Retryable reports whether trying the same transaction again might work: only a
// timeout is worth retrying. The rest fail the same way every time.
func Retryable(err error) bool {
	return errors.Is(err, ErrTimeout)
}
