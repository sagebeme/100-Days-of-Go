# Day 16: Payment Methods

## The brief

Mama Njeri's shop in Kasarani takes M-Pesa, bank cards and cash. Her nephew wrote a checkout with one `if` for each, and every time a new way to pay arrives (a supermarket voucher, a new bank) he has to change the checkout again.

Fix it with an interface. A payment method is anything that can tell you its name, its fee and pay. The checkout works with any of them, including ones nobody has written yet: it picks the cheapest one that can afford the bill.

```
wallet := &Wallet{Owner: "Amina", Balance: 1000}
wallet.Fee(300)                       →  7
wallet.Pay(300)                       →  nil   (Balance is now 693)
(&Card{Bank: "Equity"}).Fee(2500)     →  50    (2%)

cash := &Cash{InPocket: 500}
card := &Card{Bank: "Equity", Limit: 5000}
Checkout([]Method{cash, wallet, card}, 800)
  →  "M-Pesa", nil    (cash is free but short; M-Pesa's fee is 13, the card's is 16)

Describe(&Card{Bank: "Equity", Limit: 5000, Spent: 3000})
  →  "Equity card: Ksh2000 of Ksh5000 left"
```

*The M-Pesa fees in this project are a simplified, made-up tariff, not Safaricom's real price list.*

## What you'll use

New today: **interfaces**.

- An interface lists methods. Any type that has all of them **satisfies** the interface. There is no `implements` keyword: Go checks it for you.

```go
type Method interface {
	Name() string
	Fee(amount int) int
	Pay(amount int) error
}
```

- A variable of interface type can hold any value that satisfies it. `[]Method{&Cash{}, &Wallet{}}` is a slice of different types that all pay.
- **Pointer receivers matter here.** Today's `Pay` methods have `(w *Wallet)` receivers, because paying changes the balance. That means `*Wallet` is a `Method` and plain `Wallet` is not. Write `&Wallet{...}`, not `Wallet{...}`, or you'll see `Wallet does not implement Method (method Fee has pointer receiver)`.
- `var _ Method = (*Wallet)(nil)` is a line that does nothing at run time. It makes the compiler check that `*Wallet` satisfies `Method`, so a typo in a method name shows up right where the type is defined.
- A **type switch** looks at what's inside an interface value, when you really need the concrete type's fields:

```go
switch v := m.(type) {
case *Wallet:
	fmt.Println(v.Balance) // v is a *Wallet here
case *Cash:
	fmt.Println(v.InPocket)
default:
	fmt.Println(m.Name())
}
```

Reach for a type switch rarely. Most code should only call the interface's methods, which is what lets new types plug in. The test file proves it: it defines a `voucher` type the package has never seen, and `Checkout` pays with it.

- `sort.SliceStable` sorts but keeps equal elements in their original order. You need that for "same fee, first one listed wins".

## Steps

1. Open `starter/payments.go`. The interface and the three types are written. Each type has three methods with empty bodies.
2. Write the three `Name` methods. The card's is the bank name plus ` card`.
3. Write the three `Fee` methods. The wallet's is a `switch` with no value: `case amount <= 100: return 0`, and so on down the tariff. The card's is `amount * 2 / 100`. Integer division rounds down, which is what you want.
4. Write the three `Pay` methods. Check `amount < 1` first. Work out the cost with the fee, return an error if it doesn't fit, and only then change the balance.
5. Write `Total`. It takes a `Method`, so it works for every type: `amount + m.Fee(amount)`.
6. Write `Checkout`. Copy the slice (so you don't reorder the caller's), sort the copy with `sort.SliceStable` by `Fee(amount)`, then try `Pay` on each until one returns `nil`. Return that method's `Name`. Handle `amount < 1` and "nobody could pay" with errors.
7. Write `Describe` with a type switch. The cases are `*Wallet`, `*Card` and `*Cash`, with a `default` that returns `m.Name()`.
8. Run the tests from the repo root:

```bash
go test ./phase-2-modelling/day-016-payment-methods/starter
```

## The kata: Payment Router

The same idea at till scale: a list of methods, thousands of payments, and for each one the cheapest method that can afford it. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 16
```

## When you're stuck

- **`cannot use Wallet{…} (value of struct type Wallet) as Method value: Wallet does not implement Method (method Fee has pointer receiver)`**: use `&Wallet{...}`. The methods belong to `*Wallet`.
- **`*Wallet does not implement Method (missing method Pay)`**: a method name or signature doesn't match the interface exactly. `Pay(amount int) error` must be spelled and typed the same.
- **The balance doesn't change after `Pay`**: the receiver is a value, so it changed a copy. Use `(w *Wallet)`.
- **`TestCheckout` picks the card when it wanted M-Pesa for Ksh40**: both fees are 0, so the order given decides. Use `sort.SliceStable`, not `sort.Slice`.
- **`TestCheckoutPaysOnce` shows two methods paid**: stop at the first `Pay` that returns `nil`. Don't call `Pay` to "test" a method and then call it again.
- **`impossible type switch case`**: the case type doesn't satisfy `Method`. Write `case *Wallet:`, not `case Wallet:`.
- **Still stuck?** Open `solution/payments.go`, read it once, close it, then write your own version from memory.
