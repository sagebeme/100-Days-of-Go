// Package saveload keeps a chama's contribution ledger in a JSON file, so it survives
// after the program exits.
package saveload

// Contribution is one payment into the chama.
//
// Add JSON struct tags to these types so the file uses the names in the README:
// date, amount_ksh, name, phone, contributions, notes (left out when empty), chama and
// members. PIN must never reach the file.
type Contribution struct {
	Date   string // YYYY-MM-DD
	Amount int    // whole shillings
}

// Member is one member of the chama. PIN is only for this program's memory: it is
// never written to the file.
type Member struct {
	Name          string
	Phone         string
	Contributions []Contribution
	Notes         string
	PIN           string
}

// Ledger is everything the file holds.
type Ledger struct {
	Chama   string
	Members []Member
}

// Contribute records that the member called name paid amount shillings on date. It
// returns an error if there's no such member, amount is below 1, or date isn't a real
// date in the form YYYY-MM-DD.
func (l *Ledger) Contribute(name, date string, amount int) error {
	// Your code here.
	return nil
}

// Total returns how much the member called name has paid in, or 0 if there's no such
// member.
func (l Ledger) Total(name string) int {
	// Your code here.
	return 0
}

// Save writes l to path as JSON indented with two spaces, ending with a newline.
// Empty member and contribution lists are written as [] rather than null.
//
// It writes to a temporary file in the same folder first and then renames it over
// path, so a crash halfway through never leaves a half-written ledger. The file is
// readable and writable by its owner only (0600), because it holds phone numbers.
func Save(path string, l Ledger) error {
	// Your code here.
	return nil
}

// Load reads a ledger that Save wrote. It returns an error if the file can't be read
// (a missing file's error wraps fs.ErrNotExist, so callers can check with errors.Is),
// isn't valid JSON, or has a field the Ledger doesn't know, such as a typo like
// "ammount_ksh". Every error starts with the path.
func Load(path string) (Ledger, error) {
	// Your code here.
	return Ledger{}, nil
}

// LoadOrNew loads the ledger at path. If there's no file yet, it returns a new, empty
// ledger for the chama called chama, with Members set to an empty slice. Any other
// error is returned.
func LoadOrNew(path, chama string) (Ledger, error) {
	// Your code here.
	return Ledger{}, nil
}
