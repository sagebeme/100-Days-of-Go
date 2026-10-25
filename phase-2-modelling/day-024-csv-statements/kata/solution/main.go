package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// cents reads "-1,235.50", "0.5" or "" as cents.
func cents(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0
	}
	sign := 1
	if strings.HasPrefix(s, "-") {
		sign, s = -1, s[1:]
	}
	whole, frac, _ := strings.Cut(s, ".")
	frac = (frac + "00")[:2]
	w, _ := strconv.Atoi(whole)
	f, _ := strconv.Atoi(frac)
	return sign * (w*100 + f)
}

func main() {
	in := bufio.NewReader(os.Stdin)
	name, _ := in.ReadString('\n')
	name = strings.TrimRight(name, "\r\n")

	records, err := csv.NewReader(in).ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	col := -1
	for i, h := range records[0] {
		if h == name {
			col = i
			break
		}
	}
	if col == -1 {
		fmt.Printf("no column %s\n", name)
		return
	}

	total := 0
	for _, r := range records[1:] {
		total += cents(r[col])
	}
	sign := ""
	if total < 0 {
		sign, total = "-", -total
	}
	fmt.Printf("%d rows\n%s%d.%02d\n", len(records)-1, sign, total/100, total%100)
}
