# Kata: Till Totals

A shop exports its Lipa na M-Pesa payments as JSON: an array of objects, one per payment. Add up the payments for each till number, leaving out reversed ones, and print the tills from the biggest total to the smallest.

Each object has:

- `"till"`: the till number, a string
- `"amount_ksh"`: the amount, a whole number (0 ≤ amount ≤ 10^9)
- `"reversed"`: `true` if the payment was reversed. It may be missing, which means `false`.

Objects may have other fields, like `"phone"`. Ignore them. Fields can come in any order, and the JSON can be spread over lines in any way.

**Input**

One JSON array of up to 100 000 objects.

**Output**

For each till that has at least one payment that wasn't reversed, one line: the till, its total, and its number of payments that weren't reversed. Biggest total first; equal totals in alphabetical order of the till. If there's nothing to print, print `no payments`.

**Example**

```
Input
[
  {"till": "522533", "amount_ksh": 450, "phone": "0712345678"},
  {"till": "174379", "amount_ksh": 1200},
  {"till": "522533", "amount_ksh": 300, "reversed": false},
  {"till": "174379", "amount_ksh": 5000, "reversed": true}
]

Output
174379 1200 1
522533 750 2
```

**Limits:** 2 seconds per test case.

**Go tip**

Let `encoding/json` do the parsing. A struct with tags matches the field names, a missing `reversed` gets its zero value `false`, and fields the struct doesn't have are ignored:

```go
type payment struct {
	Till     string `json:"till"`
	Amount   int    `json:"amount_ksh"`
	Reversed bool   `json:"reversed"`
}

var payments []payment
err := json.NewDecoder(os.Stdin).Decode(&payments)
```

Judge it from the repo root with `go run ./tools/kata 23`.
