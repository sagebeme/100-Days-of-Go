package jobboard

import "net/http"

// MaxBodyBytes caps a request body, JSON or form.
const MaxBodyBytes = 64 << 10

// NewServer returns the job board's handler:
//
//	GET  /api/jobs?q=&page=   a Page of search results as JSON (400 for a bad page)
//	GET  /api/jobs/{id}       one job as JSON, or 404
//	POST /api/jobs            post a job as JSON: 201 with Location, or 400, 413, 415, 422
//	GET  /                    the search page as HTML, ?q= and ?page= as above
//	GET  /jobs/{id}           one job as HTML, or 404
//	GET  /jobs/new            the form to post a job
//	POST /jobs                the form's target: 303 to the new job, or 422 with the form again
//
// JSON errors look like {"error": "..."}; a 422 adds "fields" with the ValidationError.
// Render HTML with html/template. Wrap the whole mux in http.CrossOriginProtection,
// so a cross-site POST gets 403.
func NewServer(s *Store) http.Handler {
	mux := http.NewServeMux()
	// Your code here
	return mux
}
