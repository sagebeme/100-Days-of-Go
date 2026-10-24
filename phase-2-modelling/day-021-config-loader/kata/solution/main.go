package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func expand(v string, env map[string]string) string {
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		if strings.HasPrefix(v[i:], "${") {
			if end := strings.IndexByte(v[i+2:], '}'); end != -1 {
				b.WriteString(env[v[i+2:i+2+end]])
				i += 2 + end
				continue
			}
		}
		b.WriteByte(v[i])
	}
	return b.String()
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	env := map[string]string{}
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		env[strings.TrimSpace(key)] = expand(strings.TrimSpace(value), env)
	}

	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(out, "%s=%s\n", k, env[k])
	}
}
