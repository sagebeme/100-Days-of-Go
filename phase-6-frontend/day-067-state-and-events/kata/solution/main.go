package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// change remembers what one action did, so undo can reverse it.
type change struct {
	added bool
	index int
	name  string
}

func main() {
	in := bufio.NewReader(os.Stdin)
	var n int
	fmt.Fscan(in, &n)
	var queue []string
	var history []change
	for i := 0; i < n; i++ {
		var action string
		fmt.Fscan(in, &action)
		switch action {
		case "add":
			var name string
			fmt.Fscan(in, &name)
			queue = append(queue, name)
			history = append(history, change{added: true})
		case "remove":
			var name string
			fmt.Fscan(in, &name)
			if at := slices.Index(queue, name); at >= 0 {
				queue = slices.Delete(queue, at, at+1)
				history = append(history, change{index: at, name: name})
			}
		case "undo":
			if len(history) == 0 {
				continue
			}
			last := history[len(history)-1]
			history = history[:len(history)-1]
			if last.added {
				queue = queue[:len(queue)-1]
			} else {
				queue = slices.Insert(queue, last.index, last.name)
			}
		}
	}
	if len(queue) == 0 {
		fmt.Println("empty")
		return
	}
	fmt.Println(strings.Join(queue, " "))
}
