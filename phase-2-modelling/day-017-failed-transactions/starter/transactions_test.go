package failedtransactions

import (
	"errors"
	"fmt"
	"maps"
	"testing"
)

func TestLimitErrorMessage(t *testing.T) {
	e := &LimitError{Limit: 10000, Sent: 9000, Amount: 3000}
	want := "Ksh3000 would take today's total to Ksh12000, over the Ksh10000 daily limit"
	if got := e.Error(); got != want {
		t.Errorf("LimitError.Error()\n got: %q\nwant: %q", got, want)
	}
}

func TestTxError(t *testing.T) {
	e := &TxError{Code: "QFT1ABC2DE", Err: ErrInsufficientFunds}
	if got, want := e.Error(), "transaction QFT1ABC2DE: insufficient funds"; got != want {
		t.Errorf("TxError.Error()\n got: %q\nwant: %q", got, want)
	}
	if got := e.Unwrap(); got != ErrInsufficientFunds {
		t.Errorf("TxError.Unwrap() = %v, want ErrInsufficientFunds", got)
	}
	if !errors.Is(e, ErrInsufficientFunds) {
		t.Errorf("errors.Is(TxError wrapping ErrInsufficientFunds, ErrInsufficientFunds) = false, want true")
	}
}

func newBank() *Bank {
	return &Bank{
		Balances:   map[string]int{"amina": 5000, "kip": 200, "duka": 0},
		DailyLimit: 10000,
	}
}

func TestSendSucceeds(t *testing.T) {
	b := newBank()
	if err := b.Send("QFT1ABC2DE", "amina", "duka", 1200); err != nil {
		t.Fatalf("Send(amina → duka, 1200) returned an error: %v", err)
	}
	if err := b.Send("QFT1ABC2DF", "amina", "kip", 300); err != nil {
		t.Fatalf("Send(amina → kip, 300) returned an error: %v", err)
	}
	want := map[string]int{"amina": 3500, "kip": 500, "duka": 1200}
	if !maps.Equal(b.Balances, want) {
		t.Errorf("balances after two payments\n got: %v\nwant: %v", b.Balances, want)
	}
	if got := b.SentToday["amina"]; got != 1500 {
		t.Errorf("SentToday[amina] = %d, want 1500", got)
	}
}

func TestSendErrors(t *testing.T) {
	tests := []struct {
		name     string
		from, to string
		amount   int
		want     error // the sentinel inside, or nil when the reason is a *LimitError
	}{
		{"zero", "amina", "kip", 0, ErrInvalidAmount},
		{"negative", "amina", "kip", -50, ErrInvalidAmount},
		{"unknown sender", "wanjiku", "kip", 100, ErrUnknownAccount},
		{"unknown receiver", "amina", "wanjiku", 100, ErrUnknownAccount},
		{"bad amount is checked before the account", "wanjiku", "kip", 0, ErrInvalidAmount},
		{"too little money", "kip", "duka", 201, ErrInsufficientFunds},
		{"empty account", "duka", "kip", 1, ErrInsufficientFunds},
		{"over the daily limit", "amina", "kip", 10001, nil},
	}
	for _, tt := range tests {
		b := newBank()
		before := maps.Clone(b.Balances)
		err := b.Send("RKD4XY7Z1P", tt.from, tt.to, tt.amount)
		if err == nil {
			t.Errorf("%s: Send(%s → %s, %d) = nil, want an error", tt.name, tt.from, tt.to, tt.amount)
			continue
		}
		var tx *TxError
		if !errors.As(err, &tx) || tx.Code != "RKD4XY7Z1P" {
			t.Errorf("%s: Send returned %#v, want a *TxError with Code RKD4XY7Z1P", tt.name, err)
		}
		if tt.want != nil && !errors.Is(err, tt.want) {
			t.Errorf("%s: Send returned %q, want it to wrap %q", tt.name, err, tt.want)
		}
		var limit *LimitError
		if tt.want == nil && !errors.As(err, &limit) {
			t.Errorf("%s: Send returned %q, want it to wrap a *LimitError", tt.name, err)
		}
		if !maps.Equal(b.Balances, before) {
			t.Errorf("%s: a failed Send changed the balances\n got: %v\nwant: %v", tt.name, b.Balances, before)
		}
	}
}

func TestDailyLimitAddsUp(t *testing.T) {
	b := &Bank{Balances: map[string]int{"amina": 50000, "kip": 0}, DailyLimit: 10000}
	for _, amount := range []int{4000, 5000} {
		if err := b.Send("QFT0000001", "amina", "kip", amount); err != nil {
			t.Fatalf("Send(%d) returned an error: %v", amount, err)
		}
	}
	err := b.Send("QFT0000002", "amina", "kip", 3000)
	var limit *LimitError
	if !errors.As(err, &limit) {
		t.Fatalf("third Send = %v, want a *LimitError", err)
	}
	if got, want := *limit, (LimitError{Limit: 10000, Sent: 9000, Amount: 3000}); got != want {
		t.Errorf("LimitError\n got: %+v\nwant: %+v", got, want)
	}
	if err := b.Send("QFT0000003", "amina", "kip", 1000); err != nil {
		t.Errorf("Send(1000) to reach exactly the limit = %v, want nil", err)
	}

	noLimit := &Bank{Balances: map[string]int{"amina": 50000, "kip": 0}}
	if err := noLimit.Send("QFT0000004", "amina", "kip", 40000); err != nil {
		t.Errorf("Send with DailyLimit 0 (no limit) = %v, want nil", err)
	}
}

func TestExplain(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{nil, "Confirmed."},
		{&TxError{Code: "QFT1", Err: fmt.Errorf("%w: Ksh0", ErrInvalidAmount)}, "QFT1 Failed. Enter an amount of at least Ksh1."},
		{&TxError{Code: "QFT2", Err: fmt.Errorf("%w: wanjiku", ErrUnknownAccount)}, "QFT2 Failed. Check the number and try again."},
		{&TxError{Code: "QFT3", Err: &LimitError{Limit: 10000, Sent: 3000, Amount: 8000}}, "QFT3 Failed. Your daily limit is Ksh10000. You can send Ksh7000 more today."},
		{&TxError{Code: "QFT4", Err: fmt.Errorf("%w: balance Ksh200", ErrInsufficientFunds)}, "QFT4 Failed. You do not have enough money to send this."},
		{&TxError{Code: "QFT5", Err: fmt.Errorf("calling the switch: %w", ErrTimeout)}, "QFT5 Failed. Network problem, please try again."},
		{&TxError{Code: "QFT6", Err: errors.New("disk on fire")}, "QFT6 Failed. Something went wrong."},
		{ErrInsufficientFunds, "Failed. You do not have enough money to send this."},
		{fmt.Errorf("retry 3: %w", &TxError{Code: "QFT7", Err: ErrTimeout}), "QFT7 Failed. Network problem, please try again."},
		{errors.New("something else"), "Failed. Something went wrong."},
	}
	for _, tt := range tests {
		if got := Explain(tt.err); got != tt.want {
			t.Errorf("Explain(%v)\n got: %q\nwant: %q", tt.err, got, tt.want)
		}
	}
}

func TestExplainSend(t *testing.T) {
	b := newBank()
	got := Explain(b.Send("SAB0000001", "kip", "duka", 500))
	if want := "SAB0000001 Failed. You do not have enough money to send this."; got != want {
		t.Errorf("Explain(Send(kip → duka, 500))\n got: %q\nwant: %q", got, want)
	}
}

func TestRetryable(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{ErrTimeout, true},
		{&TxError{Code: "QFT1", Err: fmt.Errorf("calling the switch: %w", ErrTimeout)}, true},
		{&TxError{Code: "QFT2", Err: ErrInsufficientFunds}, false},
		{&TxError{Code: "QFT3", Err: &LimitError{Limit: 100, Amount: 200}}, false},
		{errors.New("network timeout"), false}, // same text, different error
		{nil, false},
	}
	for _, tt := range tests {
		if got := Retryable(tt.err); got != tt.want {
			t.Errorf("Retryable(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}
