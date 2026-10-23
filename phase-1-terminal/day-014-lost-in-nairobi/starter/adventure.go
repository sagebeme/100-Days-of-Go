// Package lostinnairobi is the engine of a text adventure set in Nairobi's CBD. The map
// is plain data: one map of room descriptions and one map of exits. The engine reads a
// command, moves the player and says what they see.
package lostinnairobi

// Start is the room where every game begins.
const Start = "archives"

// Goal is the room the player is trying to reach.
const Goal = "kencom"

// Help is the text for the saidia command.
const Help = "Amri: angalia, nenda <njia>, saidia. Njia ni kaskazini, kusini, mashariki, magharibi (au n, s, e, w)."

// descriptions holds what the player sees in each room. The keys are the room names.
var descriptions = map[string]string{
	"archives":   "Uko nje ya Archives kwa Moi Avenue. Makanga wanapiga kelele, \"Rongai, Rongai!\" Hii si stage yako.",
	"odeon":      "Uko Odeon. Mat za Githurai zimejaa, na music iko loud mpaka unaiskia kwa kifua.",
	"river-road": "Uko River Road. Kuna maduka ya spare parts na jamaa anauza smokie pasua. Hapa ni rahisi kupotea.",
	"hilton":     "Uko Hilton. Watu wa ofisi wanatembea fast na simu kwa sikio.",
	"kimathi":    "Uko Kimathi Street, karibu na sanamu ya Dedan Kimathi. Ni poa, lakini hakuna mat hapa.",
	"kencom":     "Uko Kencom! Mat yako ya kwenda home iko hapa.",
}

// exits says where each way out of a room leads: exits["archives"]["kusini"] is "hilton".
// Some streets only go one way, like real CBD traffic.
var exits = map[string]map[string]string{
	"archives":   {"kaskazini": "odeon", "kusini": "hilton", "mashariki": "river-road"},
	"odeon":      {"kusini": "archives", "mashariki": "river-road"},
	"river-road": {"magharibi": "odeon", "kusini": "archives"},
	"hilton":     {"kaskazini": "archives", "kusini": "kencom", "magharibi": "kimathi"},
	"kimathi":    {"mashariki": "hilton"},
	"kencom":     {"kaskazini": "hilton"},
}

// Describe returns the description of room followed by a line listing its exits in
// alphabetical order. An unknown room gives "".
//
//	Describe("kimathi") → "Uko Kimathi Street, ... hakuna mat hapa.\nNjia: mashariki"
func Describe(room string) string {
	// Your code here. The README lists what the game must do.
	return ""
}

// Won reports whether room is the goal.
func Won(room string) bool {
	// Your code here.
	return false
}

// Step runs one command typed in room. It returns the room the player is in afterwards
// and the text to show them. Commands ignore case and extra spaces:
//
//	angalia            describe the room again
//	saidia             show Help
//	nenda <njia>       walk that way; a direction on its own works too
//
// A move that has no exit, or a command the game doesn't know, leaves the player
// where they are with a message saying why.
func Step(room, command string) (string, string) {
	// Your code here.
	return "", ""
}
