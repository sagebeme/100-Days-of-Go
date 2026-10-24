package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type header struct {
	Alg string `json:"alg"`
}

type payload struct {
	Exp *int64 `json:"exp"`
	Iat *int64 `json:"iat"`
	Nbf *int64 `json:"nbf"`
}

// decode base64url-decodes part (no padding) and unmarshals the JSON object into v.
func decode(part string, v any) bool {
	b, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return false
	}
	return json.Unmarshal(b, v) == nil
}

func check(now, skew int64, token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "MALFORMED"
	}
	var h header
	var p payload
	if !decode(parts[0], &h) || !decode(parts[1], &p) || p.Exp == nil {
		return "MALFORMED"
	}
	switch {
	case h.Alg != "HS256":
		return "BAD_ALG"
	case p.Iat != nil && *p.Iat > now+skew:
		return "BAD_IAT"
	case p.Nbf != nil && now+skew < *p.Nbf:
		return "NOT_YET"
	case now >= *p.Exp+skew:
		return "EXPIRED"
	}
	return "VALID"
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var n int
	var skew int64
	fmt.Fscan(in, &n, &skew)
	for i := 0; i < n; i++ {
		var now int64
		var token string
		fmt.Fscan(in, &now, &token)
		fmt.Fprintln(out, check(now, skew, token))
	}
}
