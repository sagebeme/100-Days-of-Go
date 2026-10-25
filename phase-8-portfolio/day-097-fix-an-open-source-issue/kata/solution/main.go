package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	q, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	for range q {
		sc.Scan() // an empty line is still a line
		if c, ok := parse(sc.Text()); ok {
			fmt.Fprintln(out, format(c))
		} else {
			fmt.Fprintln(out, "INVALID")
		}
	}
}

// parse follows the grammar in the README, one piece at a time.
func parse(s string) (int64, bool) {
	s = strings.Trim(s, " ")
	neg := false
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		neg, s = true, rest
	}
	if len(s) >= 3 && (strings.EqualFold(s[:3], "ksh") || strings.EqualFold(s[:3], "kes")) {
		s = strings.TrimLeft(s[3:], " ")
	}
	if rest, ok := strings.CutPrefix(s, "-"); ok {
		if neg {
			return 0, false
		}
		neg, s = true, rest
	}
	whole, frac, hasPoint := strings.Cut(s, ".")
	digits, ok := ungroup(whole)
	if !ok || len(digits) > 13 {
		return 0, false
	}
	if hasPoint && (len(frac) < 1 || len(frac) > 2 || !allDigits(frac)) {
		return 0, false
	}
	if len(frac) == 1 {
		frac += "0" // a tenth of a shilling is ten cents
	}
	shillings, _ := strconv.ParseInt(digits, 10, 64)
	cents, _ := strconv.ParseInt("0"+frac, 10, 64)
	total := shillings*100 + cents
	if neg {
		total = -total
	}
	return total, true
}

func ungroup(s string) (string, bool) {
	if !strings.Contains(s, ",") {
		return s, s != "" && allDigits(s)
	}
	groups := strings.Split(s, ",")
	if len(groups[0]) < 1 || len(groups[0]) > 3 || !allDigits(groups[0]) {
		return "", false
	}
	for _, g := range groups[1:] {
		if len(g) != 3 || !allDigits(g) {
			return "", false
		}
	}
	return strings.Join(groups, ""), true
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func format(c int64) string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	sh := strconv.FormatInt(c/100, 10)
	var b strings.Builder
	for i := range len(sh) {
		if i > 0 && (len(sh)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(sh[i])
	}
	return fmt.Sprintf("%sKsh %s.%02d", sign, b.String(), c%100)
}
