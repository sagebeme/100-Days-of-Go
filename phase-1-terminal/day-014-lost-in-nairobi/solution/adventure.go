// Package lostinnairobi is the engine of a text adventure set in Nairobi's CBD. The map
// is plain data: one map of room descriptions and one map of exits. The engine reads a
// command, moves the player and says what they see.
package lostinnairobi

import (
	"sort"
	"strings"
)

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

// aliases turns every word a player may use for a direction into its full name.
var aliases = map[string]string{
	"kaskazini": "kaskazini", "n": "kaskazini",
	"kusini": "kusini", "s": "kusini",
	"mashariki": "mashariki", "e": "mashariki",
	"magharibi": "magharibi", "w": "magharibi",
}

// Describe returns the description of room followed by a line listing its exits in
// alphabetical order. An unknown room gives "".
//
//	Describe("kimathi") → "Uko Kimathi Street, ... hakuna mat hapa.\nNjia: mashariki"
func Describe(room string) string {
	desc, ok := descriptions[room]
	if !ok {
		return ""
	}
	ways := make([]string, 0, len(exits[room]))
	for way := range exits[room] {
		ways = append(ways, way)
	}
	sort.Strings(ways)
	return desc + "\nNjia: " + strings.Join(ways, ", ")
}

// Won reports whether room is the goal.
func Won(room string) bool {
	return room == Goal
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
	if _, ok := descriptions[room]; !ok {
		return room, "Hiyo place sijui."
	}
	words := strings.Fields(strings.ToLower(command))
	if len(words) == 0 {
		return room, "Sema kitu, msee."
	}

	var way string
	switch verb := words[0]; {
	case verb == "angalia":
		return room, Describe(room)
	case verb == "saidia":
		return room, Help
	case verb == "nenda":
		if len(words) < 2 {
			return room, "Nenda wapi? Sema njia."
		}
		way = words[1]
	case aliases[verb] != "":
		way = verb
	default:
		return room, "Sijakuelewa, msee. Jaribu 'saidia'."
	}

	full, ok := aliases[way]
	if !ok {
		return room, "Hiyo si njia. Jaribu kaskazini, kusini, mashariki au magharibi."
	}
	next, ok := exits[room][full]
	if !ok {
		return room, "Huwezi enda huko. Hakuna njia."
	}
	text := Describe(next)
	if Won(next) {
		text += "\nUmefika! Panda mat uende home. Umeshinda!"
	}
	return next, text
}
