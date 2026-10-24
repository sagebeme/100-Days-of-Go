// Package paymentmethods pays for things with whatever is to hand: an M-Pesa wallet, a bank
// card or cash. Each one is a Method, and the checkout doesn't care which.
package paymentmethods

// Method is anything that can pay. Any type with these three methods is a Method;
// it doesn't have to say so.
type Method interface {
	// Name is what the receipt shows, like "M-Pesa" or "Equity card".
	Name() string
	// Fee is what paying amount shillings costs on top of amount.
	Fee(amount int) int
	// Pay takes amount plus the fee, or returns an error and takes nothing.
	Pay(amount int) error
}

// Wallet is an M-Pesa wallet.
type Wallet struct {
	Owner   string
	Balance int // Ksh
}

// Name returns "M-Pesa".
func (w *Wallet) Name() string {
	// Your code here.
	return ""
}

// Fee follows a simplified tariff (made up for this course, not Safaricom's real one):
//
//	up to Ksh100        free
//	Ksh101 to 500       Ksh7
//	Ksh501 to 1,000     Ksh13
//	Ksh1,001 to 5,000   Ksh33
//	over Ksh5,000       Ksh57
func (w *Wallet) Fee(amount int) int {
	// Your code here.
	return 0
}

// Pay takes amount plus the fee from the balance. It returns an error if amount is
// below 1 or the balance is too small.
func (w *Wallet) Pay(amount int) error {
	// Your code here.
	return nil
}

// Card is a bank card with a spending limit.
type Card struct {
	Bank  string
	Limit int // Ksh
	Spent int // Ksh
}

// Name returns the bank's name and "card", like "Equity card".
func (c *Card) Name() string {
	// Your code here.
	return ""
}

// Fee is 2% of amount, rounded down.
func (c *Card) Fee(amount int) int {
	// Your code here.
	return 0
}

// Pay adds amount plus the fee to Spent. It returns an error if amount is below 1 or
// the payment would take Spent over Limit.
func (c *Card) Pay(amount int) error {
	// Your code here.
	return nil
}

// Cash is the money in your pocket.
type Cash struct {
	InPocket int // Ksh
}

// Name returns "Cash".
func (c *Cash) Name() string {
	// Your code here.
	return ""
}

// Fee is always 0.
func (c *Cash) Fee(amount int) int {
	// Your code here.
	return 0
}

// Pay takes amount from the pocket. It returns an error if amount is below 1 or there
// isn't enough cash.
func (c *Cash) Pay(amount int) error {
	// Your code here.
	return nil
}

// These lines don't run. They make the compiler check that each type is a Method.
var (
	_ Method = (*Wallet)(nil)
	_ Method = (*Card)(nil)
	_ Method = (*Cash)(nil)
)

// Total returns what paying amount with m costs altogether: amount plus m's fee.
func Total(m Method, amount int) int {
	// Your code here.
	return 0
}

// Checkout pays amount with the cheapest method that can afford it and returns that
// method's Name. Cheapest means the lowest fee for this amount; methods with the same
// fee are tried in the order given. It returns an error if amount is below 1 or no
// method can pay.
func Checkout(methods []Method, amount int) (string, error) {
	// Your code here.
	return "", nil
}

// Describe says how much a method has left, using a type switch to reach the fields
// each kind of Method has:
//
//	*Wallet  "M-Pesa for Amina: Ksh1500 balance"
//	*Card    "Equity card: Ksh2000 of Ksh5000 left"
//	*Cash    "Cash: Ksh300 in pocket"
//
// Any other Method is described by its Name alone.
func Describe(m Method) string {
	// Your code here.
	return ""
}
