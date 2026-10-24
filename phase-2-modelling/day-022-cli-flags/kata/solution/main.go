package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var bools = map[string]string{
	"1": "true", "t": "true", "T": "true", "TRUE": "true", "true": "true", "True": "true",
	"0": "false", "f": "false", "F": "false", "FALSE": "false", "false": "false", "False": "false",
}

func parseInt(v string) (string, bool) {
	digits := v
	if len(digits) > 0 && (digits[0] == '+' || digits[0] == '-') {
		digits = digits[1:]
	}
	if digits == "" {
		return "", false
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return "", false
		}
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return "", false
	}
	return strconv.FormatInt(n, 10), true
}

func parse(names []string, types, values map[string]string, words []string) ([]string, error) {
	i := 0
	for i < len(words) {
		word := words[i]
		if word == "--" {
			i++
			break
		}
		if len(word) < 2 || word[0] != '-' {
			break
		}
		rest := word[1:]
		if rest[0] == '-' {
			rest = rest[1:]
		}
		if rest == "" || rest[0] == '-' || rest[0] == '=' {
			return nil, fmt.Errorf("bad flag syntax: %s", word)
		}
		name, value, hasValue := strings.Cut(rest, "=")
		typ, ok := types[name]
		if !ok {
			return nil, fmt.Errorf("flag provided but not defined: -%s", name)
		}
		i++
		if typ == "bool" {
			if !hasValue {
				values[name] = "true"
				continue
			}
			b, ok := bools[value]
			if !ok {
				return nil, fmt.Errorf("invalid boolean value %q for -%s", value, name)
			}
			values[name] = b
			continue
		}
		if !hasValue {
			if i >= len(words) {
				return nil, fmt.Errorf("flag needs an argument: -%s", name)
			}
			value = words[i]
			i++
		}
		if typ == "int" {
			n, ok := parseInt(value)
			if !ok {
				return nil, fmt.Errorf("invalid value %q for flag -%s", value, name)
			}
			value = n
		}
		values[name] = value
	}
	return words[i:], nil
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	fmt.Fscan(in, &n)
	names := make([]string, n)
	types := map[string]string{}
	values := map[string]string{}
	for i := range names {
		var typ, def string
		fmt.Fscan(in, &names[i], &typ, &def)
		types[names[i]] = typ
		values[names[i]] = def
	}
	var m int
	fmt.Fscan(in, &m)
	words := make([]string, m)
	for i := range words {
		fmt.Fscan(in, &words[i])
	}

	args, err := parse(names, types, values, words)
	if err != nil {
		fmt.Fprintln(out, "error:", err)
		return
	}
	for _, name := range names {
		fmt.Fprintf(out, "%s=%s\n", name, values[name])
	}
	fmt.Fprint(out, "args:")
	for _, a := range args {
		fmt.Fprint(out, " ", a)
	}
	fmt.Fprintln(out)
}
