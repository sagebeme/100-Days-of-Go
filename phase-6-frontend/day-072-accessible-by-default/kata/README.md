# Kata: Contrast Check

A designer has sent the colour pairs for the new booking page: text colour, then background. Before anyone squints at them on a cheap phone in the midday sun, check each pair the way WCAG 2 does.

For a colour `#rrggbb`, take each channel as a number from 0 to 255 and divide by 255 to get `c`. Make it linear: `c / 12.92` if `c <= 0.04045`, otherwise `((c + 0.055) / 1.055) ^ 2.4`. The **relative luminance** is `L = 0.2126 R + 0.7152 G + 0.0722 B`. The **contrast ratio** of two colours is `(L_lighter + 0.05) / (L_darker + 0.05)`.

The verdict uses the unrounded ratio:

| Ratio | Verdict |
|---|---|
| 7 or more | `AAA` |
| 4.5 or more | `AA` |
| 3 or more | `AA-large` (fine for large text only) |
| less than 3 | `FAIL` |

**Input**

- Line 1: `n` (1 ≤ n ≤ 10 000)
- Next `n` lines: two colours, `#rrggbb #rrggbb`. Hex digits may be upper or lower case.

**Output**

`n` lines: the ratio rounded to two decimal places, a space, and the verdict.

**Example**

```
Input              Output
4                  4.54 AA
#767676 #ffffff    4.48 AA-large
#777777 #ffffff    21.00 AAA
#000000 #ffffff    6.25 AA
#0b6e4f #ffffff
```

`#767676` is the lightest grey that passes AA on white; one step lighter fails. Watch out: a ratio of 2.996 prints as `3.00` but is still `FAIL`.

**Limits:** 2 seconds per test case.

**Go tip:** `strconv.ParseUint(s[1:3], 16, 8)` reads one channel. `math.Pow(x, 2.4)` does the power, and `fmt.Sprintf("%.2f", r)` rounds for printing only; compare the full `float64`.

Judge it from the repo root with `go run ./tools/kata 72`.
