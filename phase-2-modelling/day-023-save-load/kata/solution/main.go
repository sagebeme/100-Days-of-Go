package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type payment struct {
	Till     string `json:"till"`
	Amount   int    `json:"amount_ksh"`
	Reversed bool   `json:"reversed"`
}

type till struct {
	number       string
	total, count int
}

func main() {
	var payments []payment
	if err := json.NewDecoder(bufio.NewReader(os.Stdin)).Decode(&payments); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	byTill := map[string]*till{}
	for _, p := range payments {
		if p.Reversed {
			continue
		}
		t, ok := byTill[p.Till]
		if !ok {
			t = &till{number: p.Till}
			byTill[p.Till] = t
		}
		t.total += p.Amount
		t.count++
	}

	list := make([]till, 0, len(byTill))
	for _, t := range byTill {
		list = append(list, *t)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].total != list[j].total {
			return list[i].total > list[j].total
		}
		return list[i].number < list[j].number
	})

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	if len(list) == 0 {
		fmt.Fprintln(out, "no payments")
		return
	}
	for _, t := range list {
		fmt.Fprintln(out, t.number, t.total, t.count)
	}
}
