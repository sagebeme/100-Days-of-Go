// Package hashtagmaker turns plain phrases into hashtags, initials and short previews.
package hashtagmaker

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Hashtag joins the words of phrase into one hashtag. Each word starts with a capital
// letter and the rest of it is lowercase. Anything that isn't a letter or a digit is
// dropped. A phrase with no letters or digits gives "".
//
//	Hashtag("nairobi tech week") → "#NairobiTechWeek"
func Hashtag(phrase string) string {
	var b strings.Builder
	for _, word := range strings.Fields(phrase) {
		first := true
		for _, r := range word {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				continue
			}
			if first {
				b.WriteRune(unicode.ToUpper(r))
				first = false
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		}
	}
	if b.Len() == 0 {
		return ""
	}
	return "#" + b.String()
}

// Initials returns the first letter of each word in name, in capitals.
//
//	Initials("amina wanjiru") → "AW"
func Initials(name string) string {
	var b strings.Builder
	for _, word := range strings.Fields(name) {
		r, _ := utf8.DecodeRuneInString(word)
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// Truncate keeps the first max characters of s and adds "..." if it cut anything off.
// It counts characters (runes), not bytes, so it never splits a letter like é.
//
//	Truncate("Habari za asubuhi", 6) → "Habari..."
func Truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max < 0 {
		max = 0
	}
	runes := []rune(s)
	return string(runes[:max]) + "..."
}
