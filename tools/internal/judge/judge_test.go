package judge

import "testing"

func TestCompare(t *testing.T) {
	cases := []struct {
		name, got, want string
		ok              bool
	}{
		{"exact", "1\n2\n", "1\n2\n", true},
		{"trailing spaces and blank lines", "1  \n2\n\n\n", "1\n2", true},
		{"windows line endings", "1\r\n2\r\n", "1\n2\n", true},
		{"wrong line", "1\n3\n", "1\n2\n", false},
		{"too short", "1\n", "1\n2\n", false},
		{"too long", "1\n2\n3\n", "1\n2\n", false},
	}
	for _, c := range cases {
		if got := Compare(c.got, c.want) == ""; got != c.ok {
			t.Errorf("%s: Compare matched=%v, want %v", c.name, got, c.ok)
		}
	}
}
