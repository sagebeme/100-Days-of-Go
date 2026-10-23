// Package hashtagmaker turns plain phrases into hashtags, initials and short previews.
package hashtagmaker

// Hashtag joins the words of phrase into one hashtag. Each word starts with a capital
// letter and the rest of it is lowercase. Anything that isn't a letter or a digit is
// dropped. A phrase with no letters or digits gives "".
//
//	Hashtag("nairobi tech week") → "#NairobiTechWeek"
func Hashtag(phrase string) string {
	// Your code here. Follow the steps in the README.
	return ""
}

// Initials returns the first letter of each word in name, in capitals.
//
//	Initials("amina wanjiru") → "AW"
func Initials(name string) string {
	// Your code here.
	return ""
}

// Truncate keeps the first max characters of s and adds "..." if it cut anything off.
// It counts characters (runes), not bytes, so it never splits a letter like é.
//
//	Truncate("Habari za asubuhi", 6) → "Habari..."
func Truncate(s string, max int) string {
	// Your code here.
	return ""
}
