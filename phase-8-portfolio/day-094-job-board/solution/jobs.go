// Package jobboard is a small job board: an in-memory store, search with simple
// ranking, pagination, and one server that speaks JSON under /api and HTML everywhere
// else.
package jobboard

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

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
	fields := make([]string, 0, len(v))
	for f := range v {
		fields = append(fields, f)
	}
	slices.Sort(fields)
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = f + ": " + v[f]
	}
	return "invalid job: " + strings.Join(parts, "; ")
}

// Limits for Validate.
const (
	MaxTags   = 5
	MaxTagLen = 20
)

// Validate checks a job a user wants to post. It trims nothing: call Normalize first.
// It returns nil or a ValidationError holding every problem, not just the first.
func Validate(j Job) error {
	errs := ValidationError{}
	if n := utf8.RuneCountInString(j.Title); n < 3 || n > 80 {
		errs["title"] = "must be 3 to 80 characters"
	}
	if n := utf8.RuneCountInString(j.Company); n < 2 || n > 60 {
		errs["company"] = "must be 2 to 60 characters"
	}
	if j.Location == "" {
		errs["location"] = "is required"
	}
	if len(j.Tags) > MaxTags {
		errs["tags"] = fmt.Sprintf("at most %d tags", MaxTags)
	} else {
		for _, t := range j.Tags {
			if !validTag(t) {
				errs["tags"] = "each tag is 1 to 20 lowercase letters, digits or hyphens"
				break
			}
		}
	}
	if j.SalaryKsh < 0 {
		errs["salary_ksh"] = "can't be negative"
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validTag(t string) bool {
	if t == "" || len(t) > MaxTagLen {
		return false
	}
	for _, r := range t {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

// Normalize trims spaces from the text fields, lower-cases the tags and drops empty
// or repeated tags. It doesn't change j's Tags slice.
func Normalize(j Job) Job {
	j.Title = strings.TrimSpace(j.Title)
	j.Company = strings.TrimSpace(j.Company)
	j.Location = strings.TrimSpace(j.Location)
	var tags []string
	for _, t := range j.Tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !slices.Contains(tags, t) {
			tags = append(tags, t)
		}
	}
	j.Tags = tags
	return j
}

// Tokenize lower-cases s and splits it into words of letters and digits. Everything
// else separates words: "Go/React Dev, Nairobi" gives [go react dev nairobi].
func Tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
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
	title, company := Tokenize(j.Title), Tokenize(j.Company)
	score := 0
	seen := map[string]bool{}
	for _, q := range query {
		if seen[q] {
			continue
		}
		seen[q] = true
		if slices.Contains(title, q) {
			score += TitleScore
		}
		if slices.Contains(j.Tags, q) {
			score += TagScore
		}
		if slices.Contains(company, q) {
			score += CompanyScore
		}
	}
	return score
}

// Search returns the jobs that score above zero for query, best first, newest (highest
// ID) first among equal scores. An empty query returns every job, newest first.
// It doesn't change jobs.
func Search(jobs []Job, query string) []Job {
	words := Tokenize(query)
	type hit struct {
		job   Job
		score int
	}
	var hits []hit
	for _, j := range jobs {
		s := 1
		if len(words) > 0 {
			s = Score(j, words)
		}
		if s > 0 {
			hits = append(hits, hit{j, s})
		}
	}
	slices.SortFunc(hits, func(a, b hit) int {
		if c := cmp.Compare(b.score, a.score); c != 0 {
			return c
		}
		return cmp.Compare(b.job.ID, a.job.ID)
	})
	out := make([]Job, len(hits))
	for i, h := range hits {
		out[i] = h.job
	}
	return out
}

// Page is one page of results.
type Page struct {
	Jobs    []Job `json:"jobs"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int   `json:"total"`
	Pages   int   `json:"pages"`
}

// HasPrev and HasNext say whether to show the links.
func (p Page) HasPrev() bool { return p.Page > 1 }

// HasNext reports whether there is a page after this one.
func (p Page) HasNext() bool { return p.Page < p.Pages }

// Paginate returns page number page (from 1) of jobs, perPage at a time. A page below 1
// counts as 1, and perPage below 1 counts as DefaultPerPage. A page past the end has
// no jobs, an empty non-nil slice, but the right Total and Pages.
func Paginate(jobs []Job, page, perPage int) Page {
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if page < 1 {
		page = 1
	}
	total := len(jobs)
	p := Page{Page: page, PerPage: perPage, Total: total, Pages: (total + perPage - 1) / perPage}
	start := min((page-1)*perPage, total)
	end := min(start+perPage, total)
	p.Jobs = slices.Clone(jobs[start:end])
	if p.Jobs == nil {
		p.Jobs = []Job{}
	}
	return p
}

// DefaultPerPage is the page size when none is given.
const DefaultPerPage = 10

// Store keeps jobs in memory. It is safe for concurrent use.
type Store struct {
	mu   sync.RWMutex
	jobs []Job
}

// NewStore returns an empty store.
func NewStore() *Store { return &Store{} }

// Add normalizes and validates j, gives it the next ID (from 1) and stores it.
func (s *Store) Add(j Job) (Job, error) {
	j = Normalize(j)
	if err := Validate(j); err != nil {
		return Job{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	j.ID = len(s.jobs) + 1
	s.jobs = append(s.jobs, j)
	return j, nil
}

// Get returns the job with the given ID.
func (s *Store) Get(id int) (Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id < 1 || id > len(s.jobs) {
		return Job{}, false
	}
	return s.jobs[id-1], true
}

// All returns a copy of every job, in the order they were added.
func (s *Store) All() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.jobs)
}
