package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var header = regexp.MustCompile(`^([a-z]+)(\([a-z0-9-]+\))?(!)?: [^ ]`)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(sc.Text()), "v"), ".")
	var v [3]int
	for i := range v {
		v[i], _ = strconv.Atoi(parts[i])
	}
	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))

	level := 0 // 0 none, 1 patch, 2 minor, 3 major
	for i := 0; i < n && sc.Scan(); i++ {
		m := header.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		switch {
		case m[3] == "!":
			level = max(level, 3)
		case m[1] == "feat":
			level = max(level, 2)
		case m[1] == "fix" || m[1] == "perf":
			level = max(level, 1)
		}
	}

	switch level {
	case 3:
		v = [3]int{v[0] + 1, 0, 0}
	case 2:
		v = [3]int{v[0], v[1] + 1, 0}
	case 1:
		v[2]++
	}
	fmt.Printf("v%d.%d.%d\n", v[0], v[1], v[2])
	fmt.Println([]string{"none", "patch", "minor", "major"}[level])
}
