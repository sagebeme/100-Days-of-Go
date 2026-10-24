package jobboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// serve sends one request to a fresh server holding jobs, and returns the response.
func serve(t *testing.T, s *Store, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	NewServer(s).ServeHTTP(w, r)
	return w
}

func seeded(t *testing.T, n int) *Store {
	t.Helper()
	s := NewStore()
	for i := 1; i <= n; i++ {
		if _, err := s.Add(Job{Title: fmt.Sprintf("Go Developer %d", i), Company: "Twende", Location: "Nairobi"}); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func postJSON(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/jobs", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestAPIList(t *testing.T) {
	s := seeded(t, 12)
	w := serve(t, s, httptest.NewRequest(http.MethodGet, "/api/jobs?q=developer&page=2", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/jobs: status %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var p Page
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatalf("body isn't a Page: %v\n%s", err, w.Body)
	}
	if p.Total != 12 || p.Page != 2 || p.Pages != 2 || len(p.Jobs) != 2 {
		t.Errorf("got total %d, page %d, pages %d, %d jobs; want 12, 2, 2, 2", p.Total, p.Page, p.Pages, len(p.Jobs))
	}
	if len(p.Jobs) == 2 && p.Jobs[0].ID != 2 {
		t.Errorf("page 2 starts with job %d, want 2 (newest first)", p.Jobs[0].ID)
	}
	for _, f := range []string{`"jobs"`, `"per_page"`, `"salary_ksh"`} {
		if !strings.Contains(w.Body.String(), f) {
			t.Errorf("JSON has no %s field: %s", f, w.Body)
		}
	}
}

func TestAPIEmptyListIsArray(t *testing.T) {
	w := serve(t, NewStore(), httptest.NewRequest(http.MethodGet, "/api/jobs?q=plumber", nil))
	if !strings.Contains(w.Body.String(), `"jobs":[]`) {
		t.Errorf("no results should give \"jobs\":[], got %s", w.Body)
	}
}

func TestAPIBadPage(t *testing.T) {
	for _, page := range []string{"abc", "0", "-1", "1.5"} {
		w := serve(t, seeded(t, 1), httptest.NewRequest(http.MethodGet, "/api/jobs?page="+page, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET /api/jobs?page=%s: status %d, want 400", page, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"error"`) {
			t.Errorf("GET /api/jobs?page=%s: body %s has no \"error\"", page, w.Body)
		}
	}
}

func TestAPIGet(t *testing.T) {
	s := seeded(t, 2)
	w := serve(t, s, httptest.NewRequest(http.MethodGet, "/api/jobs/2", nil))
	var j Job
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &j) != nil || j.ID != 2 {
		t.Errorf("GET /api/jobs/2: status %d, body %s; want 200 and job 2", w.Code, w.Body)
	}
	for _, path := range []string{"/api/jobs/3", "/api/jobs/0", "/api/jobs/abc"} {
		if w := serve(t, s, httptest.NewRequest(http.MethodGet, path, nil)); w.Code != http.StatusNotFound {
			t.Errorf("GET %s: status %d, want 404", path, w.Code)
		}
	}
}

func TestAPICreate(t *testing.T) {
	s := NewStore()
	w := serve(t, s, postJSON(`{"id": 99, "title": "Go Developer", "company": "Twende", "location": "Nairobi", "tags": ["Go"], "salary_ksh": 200000}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/jobs: status %d, want 201\n%s", w.Code, w.Body)
	}
	if loc := w.Header().Get("Location"); loc != "/api/jobs/1" {
		t.Errorf("Location = %q, want /api/jobs/1 (the server picks IDs)", loc)
	}
	if j, ok := s.Get(1); !ok || j.Tags[0] != "go" {
		t.Errorf("stored job = %+v, %v; want job 1 with tag go", j, ok)
	}
}

func TestAPICreateErrors(t *testing.T) {
	tests := []struct {
		name       string
		req        *http.Request
		wantStatus int
		wantInBody string
	}{
		{"broken JSON", postJSON(`{"title": `), http.StatusBadRequest, `"error"`},
		{"unknown field", postJSON(`{"title": "Go Dev", "company": "Twende", "location": "Nairobi", "admin": true}`), http.StatusBadRequest, `"error"`},
		{"invalid job", postJSON(`{"title": "Go", "company": "Twende"}`), http.StatusUnprocessableEntity, `"location"`},
		{"form instead of JSON", func() *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/api/jobs", strings.NewReader("title=Go"))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r
		}(), http.StatusUnsupportedMediaType, `"error"`},
		{"huge body", postJSON(`{"title": "` + strings.Repeat("a", MaxBodyBytes) + `"}`), http.StatusRequestEntityTooLarge, `"error"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore()
			w := serve(t, s, tt.req)
			if w.Code != tt.wantStatus {
				t.Errorf("status %d, want %d\n%s", w.Code, tt.wantStatus, w.Body)
			}
			if !strings.Contains(w.Body.String(), tt.wantInBody) {
				t.Errorf("body %s doesn't contain %s", w.Body, tt.wantInBody)
			}
			if n := len(s.All()); n != 0 {
				t.Errorf("store holds %d jobs after a rejected POST, want 0", n)
			}
		})
	}
}

func TestHTMLList(t *testing.T) {
	s := seeded(t, 12)
	s.Add(Job{Title: "Accountant", Company: "Chama Plus", Location: "Nakuru"})
	w := serve(t, s, httptest.NewRequest(http.MethodGet, "/?q=developer", nil))
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("GET /: status %d, Content-Type %q; want 200 and text/html", w.Code, w.Header().Get("Content-Type"))
	}
	if !strings.Contains(body, "Go Developer 12") || strings.Contains(body, "Accountant") {
		t.Errorf("page 1 of \"developer\" should list Go Developer 12 and not the Accountant:\n%s", body)
	}
	if strings.Contains(body, "Go Developer 2<") {
		t.Errorf("page 1 should stop after 10 jobs, but shows Go Developer 2:\n%s", body)
	}
	if !strings.Contains(body, `href="/jobs/12"`) {
		t.Errorf("each job should link to /jobs/{id}:\n%s", body)
	}
	if !strings.Contains(body, "page=2") {
		t.Errorf("page 1 of 2 should link to page=2:\n%s", body)
	}
}

func TestHTMLEscapes(t *testing.T) {
	s := NewStore()
	s.Add(Job{Title: `<script>alert(1)</script>`, Company: "Evil & Co", Location: "Nairobi"})
	for _, path := range []string{"/", "/jobs/1", "/?q=%3Cscript%3E"} {
		body := serve(t, s, httptest.NewRequest(http.MethodGet, path, nil)).Body.String()
		if strings.Contains(body, "<script>") {
			t.Errorf("GET %s: the page contains a raw <script>; use html/template:\n%s", path, body)
		}
	}
	body := serve(t, s, httptest.NewRequest(http.MethodGet, "/jobs/1", nil)).Body.String()
	if !strings.Contains(body, "&lt;script&gt;") || !strings.Contains(body, "Evil &amp; Co") {
		t.Errorf("GET /jobs/1 should show the title and company escaped:\n%s", body)
	}
}

func TestHTMLJob(t *testing.T) {
	s := seeded(t, 1)
	if w := serve(t, s, httptest.NewRequest(http.MethodGet, "/jobs/1", nil)); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Go Developer 1") {
		t.Errorf("GET /jobs/1: status %d, want 200 and the title", w.Code)
	}
	if w := serve(t, s, httptest.NewRequest(http.MethodGet, "/jobs/9", nil)); w.Code != http.StatusNotFound {
		t.Errorf("GET /jobs/9: status %d, want 404", w.Code)
	}
	if w := serve(t, s, httptest.NewRequest(http.MethodGet, "/jobs/new", nil)); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `name="title"`) {
		t.Errorf("GET /jobs/new: status %d, want 200 and a form with a title field", w.Code)
	}
}

func postForm(v url.Values) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(v.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func TestHTMLPostForm(t *testing.T) {
	s := NewStore()
	w := serve(t, s, postForm(url.Values{"title": {"Data Analyst"}, "company": {"Safari"}, "location": {"Mombasa"}, "tags": {"SQL, python"}, "salary_ksh": {"120000"}}))
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/jobs/1" {
		t.Fatalf("POST /jobs: status %d, Location %q; want 303 to /jobs/1\n%s", w.Code, w.Header().Get("Location"), w.Body)
	}
	if j, _ := s.Get(1); j.SalaryKsh != 120000 || len(j.Tags) != 2 || j.Tags[0] != "sql" {
		t.Errorf("stored %+v, want salary 120000 and tags [sql python]", j)
	}
}

func TestHTMLPostFormInvalid(t *testing.T) {
	s := NewStore()
	w := serve(t, s, postForm(url.Values{"title": {"Data Analyst"}, "company": {"S"}, "location": {"Mombasa"}}))
	body := w.Body.String()
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /jobs with a bad company: status %d, want 422", w.Code)
	}
	if !strings.Contains(body, `value="Data Analyst"`) {
		t.Errorf("the form should come back with what the user typed:\n%s", body)
	}
	if !strings.Contains(body, "2 to 60 characters") {
		t.Errorf("the form should show the error message for company:\n%s", body)
	}
	if w := serve(t, s, postForm(url.Values{"title": {"Data Analyst"}, "company": {"Safari"}, "location": {"Mombasa"}, "salary_ksh": {"lots"}})); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("POST /jobs with salary_ksh=lots: status %d, want 422", w.Code)
	}
	if n := len(s.All()); n != 0 {
		t.Errorf("store holds %d jobs after bad forms, want 0", n)
	}
}

func TestCrossOriginPostIsRejected(t *testing.T) {
	s := NewStore()
	form := postForm(url.Values{"title": {"Data Analyst"}, "company": {"Safari"}, "location": {"Mombasa"}})
	form.Header.Set("Sec-Fetch-Site", "cross-site")
	if w := serve(t, s, form); w.Code != http.StatusForbidden {
		t.Errorf("cross-site POST /jobs: status %d, want 403", w.Code)
	}
	api := postJSON(`{"title": "Go Developer", "company": "Twende", "location": "Nairobi"}`)
	api.Header.Set("Origin", "https://evil.example")
	if w := serve(t, s, api); w.Code != http.StatusForbidden {
		t.Errorf("POST /api/jobs from another Origin: status %d, want 403", w.Code)
	}
	if n := len(s.All()); n != 0 {
		t.Errorf("store holds %d jobs after cross-site POSTs, want 0", n)
	}
	get := httptest.NewRequest(http.MethodGet, "/", nil)
	get.Header.Set("Sec-Fetch-Site", "cross-site")
	if w := serve(t, s, get); w.Code != http.StatusOK {
		t.Errorf("cross-site GET /: status %d, want 200 (safe methods are allowed)", w.Code)
	}
}
