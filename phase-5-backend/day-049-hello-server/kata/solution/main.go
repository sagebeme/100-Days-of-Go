package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var methods = map[string]bool{
	"GET": true, "HEAD": true, "POST": true, "PUT": true,
	"PATCH": true, "DELETE": true, "OPTIONS": true,
}

func parse(line string) string {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "400"
	}
	for _, p := range parts {
		if p == "" {
			return "400"
		}
	}
	method, target, version := parts[0], parts[1], parts[2]
	if !methods[method] {
		return "400"
	}
	if !strings.HasPrefix(target, "/") {
		return "400"
	}
	if version != "HTTP/1.0" && version != "HTTP/1.1" {
		if strings.HasPrefix(version, "HTTP/") {
			return "505"
		}
		return "400"
	}
	path, query, _ := strings.Cut(target, "?")
	count := 0
	for _, piece := range strings.Split(query, "&") {
		if piece != "" {
			count++
		}
	}
	return method + " " + path + " " + strconv.Itoa(count)
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	sc.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(sc.Text()))
	for i := 0; i < n && sc.Scan(); i++ {
		fmt.Fprintln(out, parse(sc.Text()))
	}
}
