# Kata: Column Sum

Sum one column of a CSV statement. It sounds like splitting lines on commas. It isn't: a cell in double quotes can hold commas, newlines and doubled quotes (`""` means one `"`), and amounts like `"1,235.50"` are quoted precisely because they contain a comma.

**Input**

- Line 1: the name of the column to sum
- The rest: a CSV file following RFC 4180, the standard CSV rules. Its first record is the header. Every record has the same number of cells. At most 200 000 bytes.

Each cell in the column is empty (which counts as 0) or an amount: an optional `-`, digits with optional commas between thousands, and optionally `.` and one or two digits.

**Output**

- Line 1: the number of records after the header, as `N rows`
- Line 2: the sum, with a `-` if it's negative, and always two decimal places, with no commas: `-1335.50`

If no header cell is exactly the column name, print one line: `no column NAME`.

**Example**

```
Input
Withdrawn
Receipt No.,Details,Paid In,Withdrawn
SC1,Merchant Payment to 512345 - SOKO BORA,,"-1,235.50"
SC2,Airtime Purchase,,-100.00
SC3,"Funds received - ""BIG"" JOE, NAIROBI","5,000.00",

Output
3 rows
-1335.50
```

**Limits:** 2 seconds per test case.

**Go tip**

Read the first line yourself with a `bufio.Reader`, then hand the same reader to `encoding/csv`, which knows all the quoting rules:

```go
in := bufio.NewReader(os.Stdin)
name, _ := in.ReadString('\n')
name = strings.TrimRight(name, "\r\n")
records, err := csv.NewReader(in).ReadAll()
```

Add the amounts up as whole cents in an `int`, never as `float64`.

Judge it from the repo root with `go run ./tools/kata 24`.
