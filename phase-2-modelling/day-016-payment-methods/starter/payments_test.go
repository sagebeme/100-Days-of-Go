package paymentmethods

import "testing"

func TestNames(t *testing.T) {
	tests := []struct {
		m    Method
		want string
	}{
		{&Wallet{Owner: "Amina"}, "M-Pesa"},
		{&Card{Bank: "Equity"}, "Equity card"},
		{&Card{Bank: "KCB"}, "KCB card"},
		{&Cash{}, "Cash"},
	}
	for _, tt := range tests {
		if got := tt.m.Name(); got != tt.want {
			t.Errorf("%#v.Name() = %q, want %q", tt.m, got, tt.want)
		}
	}
}

func TestFees(t *testing.T) {
	w, c, cash := &Wallet{}, &Card{Bank: "Equity"}, &Cash{}
	tests := []struct {
		m      Method
		amount int
		want   int
	}{
		{w, 1, 0},
		{w, 100, 0},
		{w, 101, 7},
		{w, 500, 7},
		{w, 501, 13},
		{w, 1000, 13},
		{w, 1001, 33},
		{w, 5000, 33},
		{w, 5001, 57},
		{w, 70000, 57},
		{c, 100, 2},
		{c, 149, 2},
		{c, 49, 0},
		{c, 2500, 50},
		{cash, 2500, 0},
	}
	for _, tt := range tests {
		if got := tt.m.Fee(tt.amount); got != tt.want {
			t.Errorf("%s Fee(%d) = %d, want %d", tt.m.Name(), tt.amount, got, tt.want)
		}
	}
}

func TestTotal(t *testing.T) {
	if got := Total(&Wallet{}, 800); got != 813 {
		t.Errorf("Total(wallet, 800) = %d, want 813", got)
	}
	if got := Total(&Card{}, 800); got != 816 {
		t.Errorf("Total(card, 800) = %d, want 816", got)
	}
	if got := Total(&Cash{}, 800); got != 800 {
		t.Errorf("Total(cash, 800) = %d, want 800", got)
	}
}

func TestWalletPay(t *testing.T) {
	w := &Wallet{Owner: "Amina", Balance: 1000}
	if err := w.Pay(300); err != nil {
		t.Fatalf("Pay(300) with Ksh1000 returned an error: %v", err)
	}
	if w.Balance != 693 {
		t.Errorf("after Pay(300), Balance = %d, want 693 (300 plus a Ksh7 fee)", w.Balance)
	}
	for _, amount := range []int{690, 0, -5} {
		if err := w.Pay(amount); err == nil {
			t.Errorf("Pay(%d) with Ksh693 = nil, want an error", amount)
		}
	}
	if w.Balance != 693 {
		t.Errorf("after failed payments, Balance = %d, want 693", w.Balance)
	}
	if err := w.Pay(680); err != nil || w.Balance != 0 {
		t.Errorf("Pay(680) with Ksh693 = %v and Balance %d; want nil and 0", err, w.Balance)
	}
}

func TestCardPay(t *testing.T) {
	c := &Card{Bank: "Equity", Limit: 5000, Spent: 1000}
	if err := c.Pay(2500); err != nil {
		t.Fatalf("Pay(2500) returned an error: %v", err)
	}
	if c.Spent != 3550 {
		t.Errorf("after Pay(2500), Spent = %d, want 3550", c.Spent)
	}
	if err := c.Pay(1430); err == nil {
		t.Errorf("Pay(1430) with Ksh1450 left = nil, want an error (it costs 1458 with the fee)")
	}
	if err := c.Pay(0); err == nil {
		t.Errorf("Pay(0) = nil, want an error")
	}
	if err := c.Pay(1421); err != nil || c.Spent != 4999 {
		t.Errorf("Pay(1421) = %v and Spent %d; want nil and 4999", err, c.Spent)
	}
}

func TestCashPay(t *testing.T) {
	c := &Cash{InPocket: 200}
	if err := c.Pay(150); err != nil || c.InPocket != 50 {
		t.Errorf("Pay(150) with Ksh200 = %v and InPocket %d; want nil and 50", err, c.InPocket)
	}
	if err := c.Pay(51); err == nil {
		t.Errorf("Pay(51) with Ksh50 = nil, want an error")
	}
	if err := c.Pay(-1); err == nil {
		t.Errorf("Pay(-1) = nil, want an error")
	}
	if c.InPocket != 50 {
		t.Errorf("after failed payments, InPocket = %d, want 50", c.InPocket)
	}
}

func TestCheckout(t *testing.T) {
	tests := []struct {
		name     string
		cash     int
		balance  int
		limit    int
		amount   int
		wantName string
	}{
		{"cash is free, so it goes first", 1000, 5000, 5000, 800, "Cash"},
		{"not enough cash: M-Pesa is next cheapest", 500, 5000, 5000, 800, "M-Pesa"},
		{"M-Pesa short too: the card", 500, 600, 5000, 800, "Equity card"},
		{"small amounts: card fee 0 ties with M-Pesa, M-Pesa is listed first", 0, 1000, 5000, 40, "M-Pesa"},
		{"card is cheaper than M-Pesa for Ksh300", 0, 1000, 5000, 300, "Equity card"},
	}
	for _, tt := range tests {
		cash := &Cash{InPocket: tt.cash}
		wallet := &Wallet{Owner: "Amina", Balance: tt.balance}
		card := &Card{Bank: "Equity", Limit: tt.limit}
		got, err := Checkout([]Method{cash, wallet, card}, tt.amount)
		if err != nil || got != tt.wantName {
			t.Errorf("%s: Checkout(Ksh%d)\n got: %q, %v\nwant: %q, <nil>", tt.name, tt.amount, got, err, tt.wantName)
		}
	}
}

func TestCheckoutPaysOnce(t *testing.T) {
	cash := &Cash{InPocket: 100}
	wallet := &Wallet{Balance: 2000}
	card := &Card{Bank: "KCB", Limit: 2000}
	if _, err := Checkout([]Method{card, wallet, cash}, 1000); err != nil {
		t.Fatalf("Checkout(1000) returned an error: %v", err)
	}
	// Fees for Ksh1000: cash 0 (too little), M-Pesa 13, card 20. M-Pesa pays.
	if cash.InPocket != 100 || wallet.Balance != 987 || card.Spent != 0 {
		t.Errorf("after Checkout(1000): cash %d, M-Pesa %d, card spent %d; want 100, 987, 0",
			cash.InPocket, wallet.Balance, card.Spent)
	}
}

func TestCheckoutFails(t *testing.T) {
	methods := []Method{&Cash{InPocket: 50}, &Wallet{Balance: 60}}
	if got, err := Checkout(methods, 100); err == nil {
		t.Errorf("Checkout(100) with Ksh50 cash and Ksh60 M-Pesa = %q, nil; want an error", got)
	}
	if got, err := Checkout(methods, 0); err == nil {
		t.Errorf("Checkout(0) = %q, nil; want an error", got)
	}
	if got, err := Checkout(nil, 100); err == nil {
		t.Errorf("Checkout(nil, 100) = %q, nil; want an error", got)
	}
}

// voucher is a Method the payments package has never heard of. It satisfies Method
// just by having the three methods.
type voucher struct{ left int }

func (v *voucher) Name() string       { return "Naivas voucher" }
func (v *voucher) Fee(amount int) int { return 0 }
func (v *voucher) Pay(amount int) error {
	v.left -= amount
	return nil
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		m    Method
		want string
	}{
		{&Wallet{Owner: "Amina", Balance: 1500}, "M-Pesa for Amina: Ksh1500 balance"},
		{&Card{Bank: "Equity", Limit: 5000, Spent: 3000}, "Equity card: Ksh2000 of Ksh5000 left"},
		{&Cash{InPocket: 300}, "Cash: Ksh300 in pocket"},
		{&voucher{left: 1000}, "Naivas voucher"},
	}
	for _, tt := range tests {
		if got := Describe(tt.m); got != tt.want {
			t.Errorf("Describe(%#v)\n got: %q\nwant: %q", tt.m, got, tt.want)
		}
	}
}

func TestCheckoutTakesAnyMethod(t *testing.T) {
	v := &voucher{left: 1000}
	got, err := Checkout([]Method{&Wallet{Balance: 5000}, v}, 600)
	if err != nil || got != "Naivas voucher" {
		t.Errorf("Checkout with a fee-free voucher\n got: %q, %v\nwant: %q, <nil>", got, err, "Naivas voucher")
	}
}
