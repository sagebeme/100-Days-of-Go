// Package jobboard is a small job board: an in-memory store, search with simple
// ranking, pagination, and one server that speaks JSON under /api and HTML everywhere
// else.
package jobboard

// Job is one advert on the board.
type Job struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Company   string   `json:"company"`
	Location  string   `json:"location"`
	Tags      []string `json:"tags"`
	SalaryKsh int      `json:"salary_ksh"` // monthly, 0 means "not stated"
}

// ValidationError maps a field name (as in the JSON) to what's wrong with it.
type ValidationError map[string]string

func (v ValidationError) Error() string {
	// Your code here: list the fields in a fixed order.
	return "invalid job"
}

// Limits for Validate.
const (
	MaxTags   = 5
	MaxTagLen = 20
)

// Validate checks a job a user wants to post, after Normalize. It returns nil or a
// ValidationError holding every problem, not just the first:
//
//	title       3 to 80 characters (runes, not bytes)
//	company     2 to 60 characters
//	location    required
//	tags        at most MaxTags, each 1 to MaxTagLen of a-z, 0-9 and -
//	salary_ksh  not negative
//
// The company message must say "2 to 60 characters".
func Validate(j Job) error {
	// Your code here
	return nil
}

// Normalize trims spaces from the text fields, lower-cases the tags and drops empty
// or repeated tags. It doesn't change j's Tags slice.
func Normalize(j Job) Job {
	// Your code here
	return j
}

// Tokenize lower-cases s and splits it into words of letters and digits. Everything
// else separates words: "Go/React Dev, Nairobi" gives [go react dev nairobi].
func Tokenize(s string) []string {
	// Your code here
	return nil
}

// Points for each query word, by where it matches.
const (
	TitleScore   = 3
	TagScore     = 2
	CompanyScore = 1
)

// Score adds up the points for each distinct query word: TitleScore if a word of the
// title equals it, plus TagScore if a tag does, plus CompanyScore if a word of the
// company does.
func Score(j Job, query []string) int {
	// Your code here
	return 0
}

// Search returns the jobs that score above zero for query, best first, newest (highest
// ID) first among equal scores. A query with no words returns every job, newest first.
// It doesn't change jobs.
func Search(jobs []Job, query string) []Job {
	// Your code here
	return nil
}

// Page is one page of results.
type Page struct {
	Jobs    []Job `json:"jobs"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int   `json:"total"`
	Pages   int   `json:"pages"`
}

// HasPrev reports whether there is a page before this one.
func (p Page) HasPrev() bool {
	// Your code here
	return false
}

// HasNext reports whether there is a page after this one.
func (p Page) HasNext() bool {
	// Your code here
	return false
}

// DefaultPerPage is the page size when none is given.
const DefaultPerPage = 10

// Paginate returns page number page (from 1) of jobs, perPage at a time. A page below 1
// counts as 1, and perPage below 1 counts as DefaultPerPage. A page past the end has
// no jobs, an empty non-nil slice, but the right Total and Pages.
func Paginate(jobs []Job, page, perPage int) Page {
	// Your code here
	return Page{}
}

// Store keeps jobs in memory. It is safe for concurrent use.
type Store struct {
}

// NewStore returns an empty store.
func NewStore() *Store { return &Store{} }

// Add normalizes and validates j, gives it the next ID (from 1) and stores it. A
// rejected job uses up no ID.
func (s *Store) Add(j Job) (Job, error) {
	// Your code here
	return Job{}, nil
}

// Get returns the job with the given ID.
func (s *Store) Get(id int) (Job, bool) {
	// Your code here
	return Job{}, false
}

// All returns a copy of every job, in the order they were added.
func (s *Store) All() []Job {
	// Your code here
	return nil
}
