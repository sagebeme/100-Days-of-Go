# Kata: Xref Table

A PDF reader starts at the end of the file. `startxref` says where the cross-reference table is, and the table says the byte offset of every object. Get one offset wrong and the reader has to repair the file, or gives up. Build the table from the sizes of the objects.

The file is a header of `h` bytes, then objects 1 to `n` in order. Object `i` with a body of `L` bytes is written as:

```
i 0 obj\n  <L bytes of body>  \nendobj\n
```

So it takes `len("i 0 obj\n") + L + len("\nendobj\n")` bytes, and the number of digits in `i` matters. The xref table starts right after the last object.

**Input**

- Line 1: `h` and `n` (1 ≤ h ≤ 10^6, 1 ≤ n ≤ 100 000)
- Line 2: `n` numbers, the body lengths `L1 ... Ln` (0 ≤ Li ≤ 10^9)

**Output**

- `xref`
- `0 N`, where `N` is `n + 1`
- `0000000000 65535 f`
- For each object, its offset as 10 digits with leading zeros, then ` 00000 n`
- `startxref`
- The offset of the xref table, as a plain number

(A real xref entry ends with a space before the newline, to make it 20 bytes. The judge ignores trailing spaces, so you may leave it out here.)

**Example**

```
Input        Output
15 3         xref
33 43 125    0 4
             0000000000 65535 f
             0000000015 00000 n
             0000000064 00000 n
             0000000123 00000 n
             startxref
             264
```

Object 1 starts after the 15-byte header. It takes 8 + 33 + 8 = 49 bytes, so object 2 starts at 64. Object 3 starts at 64 + 8 + 43 + 8 = 123, and the table at 123 + 8 + 125 + 8 = 264.

**Limits:** 2 seconds per test case. Offsets can pass 2^31, so use `int` (64 bits) and never `int32`.

**Go tip**

`%010d` pads with zeros to 10 digits, and `len(fmt.Sprintf("%d 0 obj\n", i))` counts the header without any digit arithmetic:

```go
fmt.Fprintf(out, "%010d 00000 n\n", offset)
```

Judge it from the repo root with `go run ./tools/kata 88`.
