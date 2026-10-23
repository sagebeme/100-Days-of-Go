// Package phonecleaner turns Kenyan mobile numbers, typed any which way, into one standard form.
package phonecleaner

// Clean returns raw in the form 2547XXXXXXXX or 2541XXXXXXXX. It accepts numbers
// written as 0712 345 678, 0712-345-678, 712345678, 254712345678 or +254712345678.
// Spaces and hyphens are ignored. A plus sign is allowed only at the very start.
// Anything else gives an error.
//
//	Clean("0712 345 678") → "254712345678", nil
func Clean(raw string) (string, error) {
	// Your code here. Follow the steps in the README.
	return "", nil
}

// Mask cleans raw and hides the middle digits, the way M-Pesa shows a number on a
// receipt.
//
//	Mask("+254712345678") → "0712***678", nil
func Mask(raw string) (string, error) {
	// Your code here.
	return "", nil
}

// SameNumber reports whether a and b are the same valid phone number written two
// different ways. An invalid number is never the same as anything.
//
//	SameNumber("0712 345 678", "+254712345678") → true
func SameNumber(a, b string) bool {
	// Your code here.
	return false
}
