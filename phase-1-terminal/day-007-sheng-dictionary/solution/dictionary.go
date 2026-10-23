// Package shengdictionary is a pocket dictionary from Sheng to English.
package shengdictionary

// Lookup returns the meaning of word and true, or "" and false if dict doesn't have it.
//
//	Lookup(dict, "keja")  → "house", true
//	Lookup(dict, "mbogi") → "", false
func Lookup(dict map[string]string, word string) (string, bool) {
	meaning, ok := dict[word]
	return meaning, ok
}

// TranslateAll returns words with every word that's in dict replaced by its meaning.
// Words dict doesn't know stay as they are.
//
//	TranslateAll(dict, []string{"msee", "hii", "keja"}) → [guy hii house]
func TranslateAll(dict map[string]string, words []string) []string {
	var out []string
	for _, w := range words {
		if meaning, ok := dict[w]; ok {
			out = append(out, meaning)
		} else {
			out = append(out, w)
		}
	}
	return out
}

// WordCounts returns how many times each word appears. It returns an empty map,
// not nil, when there are no words.
//
//	WordCounts([]string{"poa", "msee", "poa"}) → map[msee:1 poa:2]
func WordCounts(words []string) map[string]int {
	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++
	}
	return counts
}

// Unknown returns the words that aren't in dict, each one once, in the order they
// first appear.
//
//	Unknown(dict, []string{"msee", "rada", "rada"}) → [rada]
func Unknown(dict map[string]string, words []string) []string {
	var unknown []string
	seen := make(map[string]bool)
	for _, w := range words {
		if _, known := dict[w]; known || seen[w] {
			continue
		}
		seen[w] = true
		unknown = append(unknown, w)
	}
	return unknown
}
