package owasptop10

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func patchMe(store ProfileStore, phone, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PATCH", "/me", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if phone != "" {
		req = req.WithContext(WithPhone(req.Context(), phone))
	}
	rec := httptest.NewRecorder()
	UpdateProfile(store).ServeHTTP(rec, req)
	return rec
}

func TestUpdateProfile(t *testing.T) {
	newStore := func() *MemProfiles {
		return NewMemProfiles(
			Profile{Phone: "0712345678", Name: "Wanjiku", Role: "passenger"},
			Profile{Phone: "0744000333", Name: "Amina", Role: "admin"},
		)
	}
	tests := []struct {
		name, phone, body string
		want              int
		wantName          string
		wantRole          string
	}{
		{"rename", "0712345678", `{"name":"Wanjiku Kamau"}`, 200, "Wanjiku Kamau", "passenger"},
		{"not logged in", "", `{"name":"x"}`, 401, "Wanjiku", "passenger"},
		{"mass assignment: role", "0712345678", `{"name":"W","role":"admin"}`, 400, "Wanjiku", "passenger"},
		{"mass assignment: phone", "0712345678", `{"phone":"0744000333","name":"pwned"}`, 400, "Wanjiku", "passenger"},
		{"empty name", "0712345678", `{"name":""}`, 422, "Wanjiku", "passenger"},
		{"name too long", "0712345678", `{"name":"` + strings.Repeat("a", 51) + `"}`, 422, "Wanjiku", "passenger"},
		{"broken JSON", "0712345678", `{"name":`, 400, "Wanjiku", "passenger"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newStore()
			rec := patchMe(store, tt.phone, tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.want, rec.Body.String())
			}
			p, _ := store.Get("0712345678")
			if p.Name != tt.wantName || p.Role != tt.wantRole {
				t.Errorf("stored profile = %+v, want name %q role %q", p, tt.wantName, tt.wantRole)
			}
			if a, _ := store.Get("0744000333"); a.Name != "Amina" {
				t.Errorf("someone else's profile changed: %+v", a)
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Errorf("body isn't JSON: %q", rec.Body.String())
			}
			if tt.want == 200 && body["name"] != tt.wantName {
				t.Errorf("response name = %v, want %q", body["name"], tt.wantName)
			}
			if tt.want != 200 && body["error"] == nil {
				t.Errorf("error response %s has no \"error\" field", rec.Body.String())
			}
			if tt.want == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("401 without WWW-Authenticate: Bearer")
			}
		})
	}
}
