package integrationtests

// This file holds test helpers. They live in a normal .go file, not a _test.go file,
// so other packages' tests could use them too, the way net/http/httptest works.

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
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
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}
	var f Fixtures
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		t.Fatalf("decoding %s: %v", path, err)
	}
	return f
}

// Seed adds f's trips to s, in order, and replaces each trip in f.Trips with the
// stored copy so the caller knows the real IDs.
func Seed(t testing.TB, s Store, f *Fixtures) {
	t.Helper()
	for i, trip := range f.Trips {
		saved, err := s.AddTrip(context.Background(), trip)
		if err != nil {
			t.Fatalf("seeding trip %d: %v", i, err)
		}
		f.Trips[i] = saved
	}
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
	t.Helper()
	f := LoadFixtures(t, "testdata/fixtures.json")
	Seed(t, store, &f)
	tokens := map[string]string{}
	for _, u := range f.Users {
		tokens[u.Token] = u.Phone
	}
	srv := httptest.NewServer((&Server{Store: store, Tokens: tokens}).Routes())
	t.Cleanup(srv.Close)
	return &Env{
		URL:      srv.URL,
		Client:   &http.Client{Timeout: 5 * time.Second},
		Store:    store,
		Fixtures: f,
	}
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
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.URL+path, r)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := e.Client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response: %v", err)
	}
	return Response{Status: resp.StatusCode, Header: resp.Header, Body: data}
}

// JSON decodes the body into v, failing the test if it isn't valid JSON.
func (r Response) JSON(t testing.TB, v any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		t.Fatalf("response body %q is not the JSON expected: %v", r.Body, err)
	}
}

// OpenTestStore returns a MemStore, unless DATABASE_URL is set. Then it creates a
// fresh Postgres schema just for this test, creates the tables in it, and returns a
// PGStore that uses it. t.Cleanup drops the schema and closes the connections.
func OpenTestStore(t testing.TB) Store {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return NewMemStore()
	}
	ctx := context.Background()
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	schema := "test_" + strings.ToLower(rand.Text())
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatalf("creating schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("dropping schema %s: %v", schema, err)
		}
		admin.Close()
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("DATABASE_URL must be a URL like postgres://user:pass@host/db: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	t.Cleanup(func() { db.Close() }) // runs before the drop: cleanups run last-in, first-out
	if _, err := db.ExecContext(ctx, Schema); err != nil {
		t.Fatalf("creating tables: %v", err)
	}
	return &PGStore{DB: db}
}
