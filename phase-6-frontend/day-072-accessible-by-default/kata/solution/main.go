package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
)

// luminance is WCAG's relative luminance of a #rrggbb colour.
func luminance(hex string) float64 {
	var lin [3]float64
	for i := range lin {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		c := float64(v) / 255
		if c <= 0.04045 {
			lin[i] = c / 12.92
		} else {
			lin[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*lin[0] + 0.7152*lin[1] + 0.0722*lin[2]
}

func verdict(r float64) string {
	switch {
	case r >= 7:
		return "AAA"
	case r >= 4.5:
		return "AA"
	case r >= 3:
		return "AA-large"
	}
	return "FAIL"
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	for i := 0; i < n; i++ {
		var a, b string
		fmt.Fscan(in, &a, &b)
		la, lb := luminance(a), luminance(b)
		r := (max(la, lb) + 0.05) / (min(la, lb) + 0.05)
		fmt.Fprintf(out, "%.2f %s\n", r, verdict(r))
	}
}
