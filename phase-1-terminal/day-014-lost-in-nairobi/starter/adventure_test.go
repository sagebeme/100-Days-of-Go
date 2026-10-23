package lostinnairobi

import "testing"

const win = "\nUmefika! Panda mat uende home. Umeshinda!"

func TestDescribe(t *testing.T) {
	tests := []struct {
		room string
		want string
	}{
		{"archives", descriptions["archives"] + "\nNjia: kaskazini, kusini, mashariki"},
		{"hilton", descriptions["hilton"] + "\nNjia: kaskazini, kusini, magharibi"},
		{"river-road", descriptions["river-road"] + "\nNjia: kusini, magharibi"},
		{"kimathi", descriptions["kimathi"] + "\nNjia: mashariki"},
		{"westlands", ""},
	}
	for _, tt := range tests {
		got := Describe(tt.room)
		if got != tt.want {
			t.Errorf("Describe(%q)\n got: %q\nwant: %q", tt.room, got, tt.want)
		}
	}
}

func TestWon(t *testing.T) {
	if !Won("kencom") {
		t.Errorf("Won(%q) = false, want true", "kencom")
	}
	if Won("archives") {
		t.Errorf("Won(%q) = true, want false", "archives")
	}
}

func TestStep(t *testing.T) {
	tests := []struct {
		room     string
		command  string
		wantRoom string
		wantText string
	}{
		// Moving.
		{"archives", "nenda kaskazini", "odeon", Describe("odeon")},
		{"archives", "nenda kusini", "hilton", Describe("hilton")},
		{"archives", "mashariki", "river-road", Describe("river-road")},
		{"archives", "n", "odeon", Describe("odeon")},
		{"odeon", "nenda e", "river-road", Describe("river-road")},
		{"hilton", "w", "kimathi", Describe("kimathi")},
		{"river-road", "  NENDA   Kusini  ", "archives", Describe("archives")},
		{"hilton", "nenda kusini", "kencom", Describe("kencom") + win},
		{"hilton", "s", "kencom", Describe("kencom") + win},

		// Looking around and asking for help.
		{"kimathi", "angalia", "kimathi", Describe("kimathi")},
		{"odeon", "ANGALIA", "odeon", Describe("odeon")},
		{"hilton", "saidia", "hilton", Help},

		// Moves that don't work.
		{"kimathi", "nenda kaskazini", "kimathi", "Huwezi enda huko. Hakuna njia."},
		{"odeon", "w", "odeon", "Huwezi enda huko. Hakuna njia."},
		{"archives", "nenda juu", "archives", "Hiyo si njia. Jaribu kaskazini, kusini, mashariki au magharibi."},
		{"archives", "nenda", "archives", "Nenda wapi? Sema njia."},

		// Commands the game doesn't know.
		{"archives", "kula smokie", "archives", "Sijakuelewa, msee. Jaribu 'saidia'."},
		{"archives", "", "archives", "Sema kitu, msee."},
		{"archives", "   ", "archives", "Sema kitu, msee."},
		{"westlands", "n", "westlands", "Hiyo place sijui."},
	}
	for _, tt := range tests {
		room, text := Step(tt.room, tt.command)
		if room != tt.wantRoom || text != tt.wantText {
			t.Errorf("Step(%q, %q)\n got: %q, %q\nwant: %q, %q", tt.room, tt.command, room, text, tt.wantRoom, tt.wantText)
		}
	}
}

func TestWalkToKencom(t *testing.T) {
	room := Start
	for _, command := range []string{"n", "e", "kusini", "angalia", "nenda kusini", "s"} {
		room, _ = Step(room, command)
	}
	if !Won(room) {
		t.Errorf("after n, e, kusini, angalia, nenda kusini, s from %q, the player is in %q, want %q", Start, room, Goal)
	}
}
