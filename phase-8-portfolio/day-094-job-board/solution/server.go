package jobboard

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// MaxBodyBytes caps a request body, JSON or form.
const MaxBodyBytes = 64 << 10

// errorBody is the shape of every JSON error. Fields is set only for a 422.
type errorBody struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// NewServer returns the job board's handler:
//
//	GET  /api/jobs?q=&page=   a Page of search results as JSON
//	GET  /api/jobs/{id}       one job as JSON
//	POST /api/jobs            post a job as JSON: 201 with Location, or 400, 415, 422
//	GET  /                    the search page as HTML
//	GET  /jobs/{id}           one job as HTML
//	GET  /jobs/new            the form to post a job
//	POST /jobs                the form's target: 303 to the new job, or 422 with the form again
//
// Every route is wrapped in http.CrossOriginProtection, so a cross-site POST gets 403.
func NewServer(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		page, ok := pageParam(w, r, true)
		if !ok {
			return
		}
		writeJSON(w, http.StatusOK, Paginate(Search(s.All(), r.URL.Query().Get("q")), page, DefaultPerPage))
	})
	mux.HandleFunc("GET /api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		job, ok := lookup(s, r)
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "no such job"})
			return
		}
		writeJSON(w, http.StatusOK, job)
	})
	mux.HandleFunc("POST /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if mt != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, errorBody{Error: "send application/json"})
			return
		}
		var in Job
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				writeJSON(w, http.StatusRequestEntityTooLarge, errorBody{Error: "body too large"})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON"})
			return
		}
		in.ID = 0 // the server picks IDs
		job, err := s.Add(in)
		var verr ValidationError
		if errors.As(err, &verr) {
			writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: "invalid job", Fields: verr})
			return
		}
		w.Header().Set("Location", "/api/jobs/"+strconv.Itoa(job.ID))
		writeJSON(w, http.StatusCreated, job)
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		page, ok := pageParam(w, r, false)
		if !ok {
			return
		}
		q := r.URL.Query().Get("q")
		render(w, http.StatusOK, "list", listData{Query: q, Page: Paginate(Search(s.All(), q), page, DefaultPerPage)})
	})
	mux.HandleFunc("GET /jobs/new", func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, "form", formData{})
	})
	mux.HandleFunc("GET /jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		job, ok := lookup(s, r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		render(w, http.StatusOK, "job", job)
	})
	mux.HandleFunc("POST /jobs", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		salary := 0
		if v := strings.TrimSpace(r.PostForm.Get("salary_ksh")); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				render(w, http.StatusUnprocessableEntity, "form", formData{Form: r.PostForm, Errors: ValidationError{"salary_ksh": "must be a whole number"}})
				return
			}
			salary = n
		}
		in := Job{
			Title:     r.PostForm.Get("title"),
			Company:   r.PostForm.Get("company"),
			Location:  r.PostForm.Get("location"),
			Tags:      strings.Split(r.PostForm.Get("tags"), ","),
			SalaryKsh: salary,
		}
		job, err := s.Add(in)
		var verr ValidationError
		if errors.As(err, &verr) {
			render(w, http.StatusUnprocessableEntity, "form", formData{Form: r.PostForm, Errors: verr})
			return
		}
		http.Redirect(w, r, "/jobs/"+strconv.Itoa(job.ID), http.StatusSeeOther)
	})

	return http.NewCrossOriginProtection().Handler(mux)
}

// pageParam reads ?page=. A missing value is page 1; anything else must be a whole
// number from 1. On a bad value it writes 400 (as JSON when api is true) and returns false.
func pageParam(w http.ResponseWriter, r *http.Request, api bool) (int, bool) {
	v := r.URL.Query().Get("page")
	if v == "" {
		return 1, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		if api {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "page must be a whole number from 1"})
		} else {
			http.Error(w, "page must be a whole number from 1", http.StatusBadRequest)
		}
		return 0, false
	}
	return n, true
}

func lookup(s *Store, r *http.Request) (Job, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return Job{}, false
	}
	return s.Get(id)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing response", "err", err)
	}
}

type listData struct {
	Query string
	Page  Page
}

type formData struct {
	Form   map[string][]string
	Errors ValidationError
}

// render runs a template into memory first, so a template error becomes a clean 500
// instead of half a page.
func render(w http.ResponseWriter, status int, name string, data any) {
	var b strings.Builder
	if err := pages.ExecuteTemplate(&b, name, data); err != nil {
		slog.Error("rendering", "template", name, "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(b.String()))
}

// The templates live in the Go file so the package is one self-contained unit. In a
// bigger app, put them in .html files and load them with embed.FS.
var pages = template.Must(template.New("pages").Funcs(template.FuncMap{
	"first": func(f map[string][]string, k string) string {
		if v := f[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	},
	"list": func(s ...string) []string { return s },
	"inc":  func(n int) int { return n + 1 },
	"dec":  func(n int) int { return n - 1 },
}).Parse(`
{{define "top"}}<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.}} · Kazi Board</title></head><body>
<header><a href="/">Kazi Board</a> · <a href="/jobs/new">Post a job</a></header><main>{{end}}

{{define "bottom"}}</main></body></html>{{end}}

{{define "list"}}{{template "top" "Jobs"}}
<form action="/" method="get" role="search">
<label for="q">Search jobs</label>
<input id="q" name="q" type="search" value="{{.Query}}">
<button>Search</button>
</form>
<p>{{.Page.Total}} jobs</p>
<ul>{{range .Page.Jobs}}
<li><a href="/jobs/{{.ID}}">{{.Title}}</a> at {{.Company}}, {{.Location}}</li>{{else}}
<li>No jobs match.</li>{{end}}
</ul>
<nav>{{if .Page.HasPrev}}<a rel="prev" href="/?q={{.Query}}&amp;page={{dec .Page.Page}}">Previous</a>{{end}}
{{if .Page.HasNext}}<a rel="next" href="/?q={{.Query}}&amp;page={{inc .Page.Page}}">Next</a>{{end}}</nav>
{{template "bottom"}}{{end}}

{{define "job"}}{{template "top" .Title}}
<h1>{{.Title}}</h1>
<p>{{.Company}}, {{.Location}}</p>
{{if .SalaryKsh}}<p>Ksh {{.SalaryKsh}} a month</p>{{end}}
<ul>{{range .Tags}}<li>{{.}}</li>{{end}}</ul>
{{template "bottom"}}{{end}}

{{define "form"}}{{template "top" "Post a job"}}
<h1>Post a job</h1>
<form action="/jobs" method="post">
{{range $f := (list "title" "company" "location" "tags" "salary_ksh")}}
<p><label for="{{$f}}">{{$f}}</label>
<input id="{{$f}}" name="{{$f}}" value="{{first $.Form $f}}">
{{with index $.Errors $f}}<strong>{{.}}</strong>{{end}}</p>{{end}}
<button>Post</button>
</form>
{{template "bottom"}}{{end}}
`))
