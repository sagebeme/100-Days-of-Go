package jobboard

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"Go Developer", []string{"go", "developer"}},
		{"Go/React Dev, Nairobi", []string{"go", "react", "dev", "nairobi"}},
		{"  Senior   SRE (Remote) ", []string{"senior", "sre", "remote"}},
		{"Mhasibu wa M-Pesa", []string{"mhasibu", "wa", "m", "pesa"}},
		{"Café Ops 24/7", []string{"café", "ops", "24", "7"}},
		{"", nil},
		{"!!!", nil},
	}
	for _, tt := range tests {
		got := Tokenize(tt.in)
		if !slices.Equal(got, tt.want) {
			t.Errorf("Tokenize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func validJob() Job {
	return Job{Title: "Go Developer", Company: "Twende Ltd", Location: "Nairobi", Tags: []string{"go", "backend"}, SalaryKsh: 250000}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		change     func(j *Job)
		wantFields []string
	}{
		{"valid", func(j *Job) {}, nil},
		{"no salary is fine", func(j *Job) { j.SalaryKsh = 0 }, nil},
		{"no tags is fine", func(j *Job) { j.Tags = nil }, nil},
		{"short title", func(j *Job) { j.Title = "Go" }, []string{"title"}},
		{"long title", func(j *Job) { j.Title = strings.Repeat("a", 81) }, []string{"title"}},
		{"80 runes of Kiswahili is fine", func(j *Job) { j.Title = strings.Repeat("é", 80) }, nil},
		{"short company", func(j *Job) { j.Company = "X" }, []string{"company"}},
		{"no location", func(j *Job) { j.Location = "" }, []string{"location"}},
		{"six tags", func(j *Job) { j.Tags = []string{"a", "b", "c", "d", "e", "f"} }, []string{"tags"}},
		{"tag with a space", func(j *Job) { j.Tags = []string{"node js"} }, []string{"tags"}},
		{"upper-case tag", func(j *Job) { j.Tags = []string{"Go"} }, []string{"tags"}},
		{"negative salary", func(j *Job) { j.SalaryKsh = -1 }, []string{"salary_ksh"}},
		{"everything wrong", func(j *Job) { *j = Job{SalaryKsh: -5} }, []string{"company", "location", "salary_ksh", "title"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := validJob()
			tt.change(&j)
			err := Validate(j)
			if tt.wantFields == nil {
				if err != nil {
					t.Fatalf("Validate = %v, want nil", err)
				}
				return
			}
			var verr ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("Validate = %v, want a ValidationError", err)
			}
			var got []string
			for f := range verr {
				got = append(got, f)
			}
			slices.Sort(got)
			if !slices.Equal(got, tt.wantFields) {
				t.Errorf("fields with errors = %v, want %v", got, tt.wantFields)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	in := Job{Title: "  Data Analyst ", Company: " Safari Data ", Location: " Mombasa", Tags: []string{"SQL", " python ", "", "sql"}}
	got := Normalize(in)
	want := Job{Title: "Data Analyst", Company: "Safari Data", Location: "Mombasa", Tags: []string{"sql", "python"}}
	if got.Title != want.Title || got.Company != want.Company || got.Location != want.Location || !slices.Equal(got.Tags, want.Tags) {
		t.Errorf("Normalize = %+v, want %+v", got, want)
	}
	if in.Tags[0] != "SQL" {
		t.Errorf("Normalize changed its argument's Tags: %v", in.Tags)
	}
}

func board() []Job {
	return []Job{
		{ID: 1, Title: "Go Developer", Company: "Twende Ltd", Location: "Nairobi", Tags: []string{"backend"}},
		{ID: 2, Title: "Backend Engineer", Company: "Pesa Go", Location: "Remote", Tags: []string{"go", "postgres"}},
		{ID: 3, Title: "React Developer", Company: "Duka Hub", Location: "Kisumu", Tags: []string{"frontend"}},
		{ID: 4, Title: "Accountant", Company: "Chama Plus", Location: "Nakuru", Tags: []string{"finance"}},
		{ID: 5, Title: "Senior Go Engineer", Company: "Safari Go", Location: "Nairobi", Tags: []string{"go"}},
	}
}

func ids(jobs []Job) []int {
	out := make([]int, len(jobs))
	for i, j := range jobs {
		out[i] = j.ID
	}
	return out
}

func TestScore(t *testing.T) {
	jobs := board()
	tests := []struct {
		job   Job
		query []string
		want  int
	}{
		{jobs[0], []string{"go"}, TitleScore},
		{jobs[1], []string{"go"}, TagScore + CompanyScore},
		{jobs[4], []string{"go"}, TitleScore + TagScore + CompanyScore},
		{jobs[4], []string{"go", "go"}, TitleScore + TagScore + CompanyScore},
		{jobs[0], []string{"developer", "backend"}, TitleScore + TagScore},
		{jobs[3], []string{"go"}, 0},
		{jobs[0], []string{"dev"}, 0}, // whole words only
	}
	for _, tt := range tests {
		got := Score(tt.job, tt.query)
		if got != tt.want {
			t.Errorf("Score(job %d, %q) = %d, want %d", tt.job.ID, tt.query, got, tt.want)
		}
	}
}

func TestSearch(t *testing.T) {
	tests := []struct {
		query string
		want  []int
	}{
		{"go", []int{5, 2, 1}},          // 6, then 3 and 3: a tie goes to the newer job
		{"Developer", []int{3, 1}},      // a tie, newest first
		{"", []int{5, 4, 3, 2, 1}},      // everything, newest first
		{"  ,, ", []int{5, 4, 3, 2, 1}}, // no words is the same as empty
		{"plumber", []int{}},            // nothing matches
		{"react frontend", []int{3}},
		{"backend go", []int{5, 2, 1}}, // 6, 6 and 5
		{"ACCOUNTANT!", []int{4}},      // case and punctuation don't matter
		{"senior go engineer", []int{5, 2, 1}},
	}
	for _, tt := range tests {
		got := ids(Search(board(), tt.query))
		if !slices.Equal(got, tt.want) {
			t.Errorf("Search(%q) = IDs %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestSearchDoesNotReorderInput(t *testing.T) {
	jobs := board()
	Search(jobs, "go")
	if got := ids(jobs); !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Errorf("Search reordered its input to %v", got)
	}
}

func TestPaginate(t *testing.T) {
	var jobs []Job
	for i := 1; i <= 23; i++ {
		jobs = append(jobs, Job{ID: i})
	}
	tests := []struct {
		page, perPage int
		wantIDs       []int
		wantPage      int
		wantPerPage   int
		wantPages     int
		prev, next    bool
	}{
		{1, 10, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 1, 10, 3, false, true},
		{3, 10, []int{21, 22, 23}, 3, 10, 3, true, false},
		{4, 10, []int{}, 4, 10, 3, true, false},
		{0, 10, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 1, 10, 3, false, true},
		{-2, 5, []int{1, 2, 3, 4, 5}, 1, 5, 5, false, true},
		{2, 0, []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 2, DefaultPerPage, 3, true, true},
		{1, 23, ids(jobs), 1, 23, 1, false, false},
	}
	for _, tt := range tests {
		p := Paginate(jobs, tt.page, tt.perPage)
		name := fmt.Sprintf("Paginate(23 jobs, %d, %d)", tt.page, tt.perPage)
		if p.Jobs == nil {
			t.Errorf("%s: Jobs is nil, want an empty slice so JSON shows []", name)
		}
		if got := ids(p.Jobs); !slices.Equal(got, tt.wantIDs) {
			t.Errorf("%s: IDs %v, want %v", name, got, tt.wantIDs)
		}
		if p.Page != tt.wantPage || p.PerPage != tt.wantPerPage || p.Total != 23 || p.Pages != tt.wantPages {
			t.Errorf("%s: page %d, per page %d, total %d, pages %d; want %d, %d, 23, %d",
				name, p.Page, p.PerPage, p.Total, p.Pages, tt.wantPage, tt.wantPerPage, tt.wantPages)
		}
		if p.HasPrev() != tt.prev || p.HasNext() != tt.next {
			t.Errorf("%s: HasPrev %v, HasNext %v; want %v, %v", name, p.HasPrev(), p.HasNext(), tt.prev, tt.next)
		}
	}
	if p := Paginate(nil, 1, 10); p.Pages != 0 || p.Total != 0 || p.Jobs == nil || p.HasNext() {
		t.Errorf("Paginate(nil, 1, 10) = %+v, want no pages, an empty slice and no next", p)
	}
}

func TestStore(t *testing.T) {
	s := NewStore()
	a, err := s.Add(Job{Title: " Go Developer ", Company: "Twende", Location: "Nairobi", Tags: []string{"GO"}})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if a.ID != 1 || a.Title != "Go Developer" || !slices.Equal(a.Tags, []string{"go"}) {
		t.Errorf("Add returned %+v, want ID 1, a trimmed title and lower-case tags", a)
	}
	if _, err := s.Add(Job{Title: "x"}); err == nil {
		t.Error("Add of an invalid job returned nil error")
	}
	b, _ := s.Add(Job{Title: "Data Analyst", Company: "Safari", Location: "Mombasa"})
	if b.ID != 2 {
		t.Errorf("second valid job has ID %d, want 2 (a rejected job uses no ID)", b.ID)
	}
	if got, ok := s.Get(2); !ok || got.Title != "Data Analyst" {
		t.Errorf("Get(2) = %+v, %v", got, ok)
	}
	for _, id := range []int{0, 3, -1} {
		if _, ok := s.Get(id); ok {
			t.Errorf("Get(%d) found a job, want none", id)
		}
	}
	if got := len(s.All()); got != 2 {
		t.Errorf("len(All()) = %d, want 2", got)
	}
}
