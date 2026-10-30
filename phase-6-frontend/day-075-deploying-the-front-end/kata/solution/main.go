package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func readManifest(in *bufio.Reader) (map[string]string, []string) {
	var n int
	fmt.Fscan(in, &n)
	hashes := make(map[string]string, n)
	paths := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var path, hash string
		fmt.Fscan(in, &path, &hash)
		hashes[path] = hash
		paths = append(paths, path)
	}
	return hashes, paths
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	live, livePaths := readManifest(in)
	next, nextPaths := readManifest(in)

	var upload, del, keep []string
	unchanged := 0
	for _, p := range nextPaths {
		if old, ok := live[p]; ok && old == next[p] {
			unchanged++
		} else {
			upload = append(upload, p)
		}
	}
	for _, p := range livePaths {
		if _, ok := next[p]; ok {
			continue
		}
		if strings.HasPrefix(p, "assets/") {
			keep = append(keep, p) // someone may still be running the old build
		} else {
			del = append(del, p)
		}
	}
	for _, group := range []struct {
		verb  string
		paths []string
	}{{"upload", upload}, {"delete", del}, {"keep", keep}} {
		sort.Strings(group.paths)
		for _, p := range group.paths {
			fmt.Fprintf(out, "%s %s\n", group.verb, p)
		}
	}
	fmt.Fprintf(out, "unchanged %d\n", unchanged)
}
