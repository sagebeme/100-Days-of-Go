# Day 17: Failed Transactions

## The brief

A payment fails and the customer gets a message: "Failed. You do not have enough money to send this." Behind that sentence the code had to know *which* failure it was, while also keeping the transaction code and the technical details for the logs.

Build both halves. `Send` moves money between accounts and returns rich errors: the transaction code wrapped around the reason. `Explain` digs into any error and picks the customer's message. `Retryable` says whether trying again could help.

```
b := &Bank{Balances: map[string]int{"amina": 5000, "kip": 200}, DailyLimit: 10000}

b.Send("QFT1ABC2DE", "amina", "kip", 1200)
  →  nil
b.Send("RKD4XY7Z1P", "kip", "amina", 5000)
  →  transaction RKD4XY7Z1P: insufficient funds: balance Ksh1400, need Ksh5000

Explain(err)    →  "RKD4XY7Z1P Failed. You do not have enough money to send this."
Retryable(err)  →  false

Explain(&TxError{Code: "QFT3", Err: &LimitError{Limit: 10000, Sent: 3000, Amount: 8000}})
  →  "QFT3 Failed. Your daily limit is Ksh10000. You can send Ksh7000 more today."
```

## What you'll use

Since Day 5 you've returned errors made with `errors.New` and `fmt.Errorf`. Today errors get structure. All of this is new:

- **Sentinel errors**: package-level values like `var ErrInsufficientFunds = errors.New("insufficient funds")`. Callers check for them by identity, not by reading the text.
- **Wrapping with `%w`**: `fmt.Errorf("%w: balance Ksh%d", ErrInsufficientFunds, bal)` makes a new error with more detail that still *contains* the sentinel. (`%v` would copy the text and lose the link.)
- **Custom error types**: any type with an `Error() string` method is an `error`. That's Day 16's interfaces at work. A struct can carry fields like `Limit` and `Sent` that a caller can read.
- **`Unwrap() error`**: a custom type that wraps another error returns it from `Unwrap`, so the standard library can look inside.
- **`errors.Is(err, target)`**: unwraps `err` layer by layer and reports whether any layer is `target`.
- **`errors.As(err, &target)`**: unwraps until it finds a layer of `target`'s type, stores it in `target`, and returns `true`.

```go
err := fmt.Errorf("paying for unga: %w", &TxError{Code: "QFT1", Err: ErrTimeout})

errors.Is(err, ErrTimeout) // true: three layers down
err == ErrTimeout          // false: == only checks the outside layer

var tx *TxError
if errors.As(err, &tx) {
	fmt.Println(tx.Code) // "QFT1"
}
```

Use `errors.Is` for "is it this particular error value", and `errors.As` for "is it this kind of error, and give me its fields".

## Why it matters at work

Error handling is where code review spends a lot of its time. Comparing error strings (`strings.Contains(err.Error(), "insufficient")`) breaks the day someone rewords a message. Comparing with `==` breaks the day someone adds context with `%w`. Showing raw errors to users leaks internals such as account names and SQL. The pattern you build today is the one production Go uses: wrap with context as the error goes up, keep the details for logs, check with `errors.Is` and `errors.As`, and turn it into a safe message at the very edge. It also decides retries. A timeout may work the second time; "insufficient funds" never will, and retrying it only hammers the system.

## Steps

1. Open `starter/transactions.go`. The sentinels and the types are there; the methods are empty.
2. Write `LimitError.Error` with `fmt.Sprintf`, matching the doc comment exactly. Then `TxError.Error` and `TxError.Unwrap`.
3. Write the checks in `Send`, in the order the doc comment gives. Wrap each sentinel with `%w` and a detail, for example `fmt.Errorf("%w: %s", ErrUnknownAccount, from)`. For the limit, return `&LimitError{...}`. A small helper function `check(from, to string, amount int) error` keeps `Send` short.
4. In `Send`, wrap whatever the checks return: `return &TxError{Code: code, Err: err}`. Only when there's no error, create `SentToday` if it's `nil` and move the money.
5. Write `Explain`. First `errors.As` for a `*TxError` to get the prefix. Then a `switch` with no value, with one `case` per reason, using `errors.Is` for the sentinels and `errors.As` for `*LimitError`.
6. Write `Retryable`: one line with `errors.Is`.
7. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-017-failed-transactions/starter
```

## The kata: Transfer Log

Process a day's transfers in order and log each one as `OK` or `FAILED` with a reason code, checked in a fixed order. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 17
```

## When you're stuck

- **`errors.Is` returns false for an error you built with `fmt.Errorf`**: you used `%v` or `%s`. Only `%w` wraps.
- **`errors.Is` returns false through a `*TxError`**: `Unwrap` is missing, misspelled, or has a value receiver while you return a pointer. It must be `func (e *TxError) Unwrap() error`.
- **`second argument to errors.As must be a non-nil pointer to either a type that implements error, or to any interface type`**: pass the address of a variable, `errors.As(err, &limit)` where `var limit *LimitError`.
- **`TestSendErrors` says a failed Send changed the balances**: run every check before you touch `Balances` or `SentToday`.
- **`assignment to entry in nil map`**: `SentToday` starts out `nil` in the tests. Make it with `map[string]int{}` before the first write.
- **The limit test fails at exactly Ksh10000**: reaching the limit is allowed; only going over it is not. Use `>`, not `>=`.
- **Still stuck?** Open `solution/transactions.go`, read it once, close it, then write your own version from memory.
