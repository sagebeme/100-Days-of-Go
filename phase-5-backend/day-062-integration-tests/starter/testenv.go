package integrationtests

// This file holds test helpers. They live in a normal .go file, not a _test.go file,
// so other packages' tests could use them too, the way net/http/httptest works.

import (
	"net/http"
	"testing"
)

// FixtureUser is a passenger the tests can log in as.
type FixtureUser struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Token string `json:"token"`
}

// Fixtures is the starting data for a test, read from testdata/fixtures.json.
type Fixtures struct {
	Trips []Trip        `json:"trips"`
	Users []FixtureUser `json:"users"`
}

// LoadFixtures reads and decodes the JSON file at path. Any problem fails the test.
func LoadFixtures(t testing.TB, path string) Fixtures {
	// Your code here
	return Fixtures{}
}

// Seed adds f's trips to s, in order, and replaces each trip in f.Trips with the
// stored copy so the caller knows the real IDs.
func Seed(t testing.TB, s Store, f *Fixtures) {
	// Your code here
}

// Env is a running API for one test.
type Env struct {
	URL      string
	Client   *http.Client
	Store    Store
	Fixtures Fixtures
}

// NewEnv loads testdata/fixtures.json, seeds store with it, and starts the whole API
// on a real local port with httptest.NewServer. The server is closed by t.Cleanup.
func NewEnv(t testing.TB, store Store) *Env {
	// Your code here
	return &Env{}
}

// Response is what came back from a request.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// Do sends a request to the running API. A non-nil body is sent as JSON with
// Content-Type: application/json; a non-empty token as "Authorization: Bearer <token>".
// Any transport error fails the test.
func (e *Env) Do(t testing.TB, method, path, token string, body any) Response {
	// Your code here
	return Response{}
}

// JSON decodes the body into v, failing the test if it isn't valid JSON.
func (r Response) JSON(t testing.TB, v any) {
	// Your code here
}

// OpenTestStore returns a MemStore, unless DATABASE_URL is set. Then it creates a
// fresh Postgres schema just for this test, creates the tables in it, and returns a
// PGStore that uses it. t.Cleanup drops the schema and closes the connections.
func OpenTestStore(t testing.TB) Store {
	// Your code here
	return NewMemStore()
}
