package cliflags

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		total, members int
		want           []int
	}{
		{12005, 5, []int{2401, 2401, 2401, 2401, 2401}},
		{10, 3, []int{4, 3, 3}},
		{11, 3, []int{4, 4, 3}},
		{7, 7, []int{1, 1, 1, 1, 1, 1, 1}},
		{500, 1, []int{500}},
	}
	for _, tt := range tests {
		got, err := Split(tt.total, tt.members)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Split(%d, %d)\n got: %v, %v\nwant: %v, <nil>", tt.total, tt.members, got, err, tt.want)
		}
	}
	for _, bad := range [][2]int{{3, 5}, {100, 0}, {100, -2}} {
		if got, err := Split(bad[0], bad[1]); err == nil {
			t.Errorf("Split(%d, %d) = %v, nil; want an error", bad[0], bad[1], got)
		}
	}
}

func TestLoan(t *testing.T) {
	tests := []struct {
		amount, rate, months     int
		interest, total, monthly int
	}{
		{20000, 10, 3, 6000, 26000, 8667},
		{10000, 10, 1, 1000, 11000, 11000},
		{5000, 0, 5, 0, 5000, 1000},
		{999, 3, 2, 59, 1058, 529},
		{1, 1, 1, 0, 1, 1},
	}
	for _, tt := range tests {
		i, tot, m := Loan(tt.amount, tt.rate, tt.months)
		if i != tt.interest || tot != tt.total || m != tt.monthly {
			t.Errorf("Loan(%d, %d, %d) = %d, %d, %d; want %d, %d, %d",
				tt.amount, tt.rate, tt.months, i, tot, m, tt.interest, tt.total, tt.monthly)
		}
	}
}

// run calls Run with args and returns the exit code and what it printed.
func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunWorks(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"split", []string{"split", "-total", "12005", "-members", "5"},
			"member 1: Ksh2,401\nmember 2: Ksh2,401\nmember 3: Ksh2,401\nmember 4: Ksh2,401\nmember 5: Ksh2,401\n"},
		{"split with = and flags in any order", []string{"split", "-members=3", "--total=10"},
			"member 1: Ksh4\nmember 2: Ksh3\nmember 3: Ksh3\n"},
		{"loan with defaults", []string{"loan", "-amount", "20000"},
			"interest: Ksh6,000\ntotal: Ksh26,000\nmonthly: Ksh8,667\n"},
		{"loan with every flag", []string{"loan", "-amount", "150000", "-rate", "5", "-months", "12"},
			"interest: Ksh90,000\ntotal: Ksh240,000\nmonthly: Ksh20,000\n"},
		{"loan at zero percent", []string{"loan", "-amount=5000", "-rate=0", "-months=5"},
			"interest: Ksh0\ntotal: Ksh5,000\nmonthly: Ksh1,000\n"},
		{"version", []string{"version"}, "chama 1.0.0\n"},
		{"help", []string{"help"}, Usage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := run(t, tt.args...)
			if code != ExitOK || stdout != tt.want || stderr != "" {
				t.Errorf("Run(%q)\n got: exit %d, stdout %q, stderr %q\nwant: exit 0, stdout %q, stderr \"\"",
					tt.args, code, stdout, stderr, tt.want)
			}
		})
	}
}

func TestRunProblems(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string // stderr must contain this
	}{
		{"no command", nil, ExitUsage, "usage: chama <command>"},
		{"unknown command", []string{"lend"}, ExitUsage, `chama: unknown command "lend"`},
		{"unknown command shows usage", []string{"lend"}, ExitUsage, "usage: chama <command>"},
		{"unknown flag", []string{"split", "-people", "5"}, ExitUsage, "flag provided but not defined: -people"},
		{"not a number", []string{"split", "-total", "lots", "-members", "5"}, ExitUsage, `invalid value "lots" for flag -total`},
		{"flag with no value", []string{"loan", "-amount"}, ExitUsage, "flag needs an argument: -amount"},
		{"missing -total", []string{"split", "-members", "5"}, ExitUsage, "chama split: -total is required"},
		{"missing -members", []string{"split", "-total", "500"}, ExitUsage, "chama split: -members is required"},
		{"missing -amount", []string{"loan", "-rate", "5"}, ExitUsage, "chama loan: -amount is required"},
		{"zero members", []string{"split", "-total", "500", "-members", "0"}, ExitUsage, "chama split: -members must be at least 1, got 0"},
		{"negative amount", []string{"loan", "-amount", "-500"}, ExitUsage, "chama loan: -amount must be at least 1, got -500"},
		{"rate too high", []string{"loan", "-amount", "500", "-rate", "101"}, ExitUsage, "chama loan: -rate must be from 0 to 100, got 101"},
		{"too many months", []string{"loan", "-amount", "500", "-months", "48"}, ExitUsage, "chama loan: -months must be from 1 to 36, got 48"},
		{"extra argument", []string{"split", "-total", "500", "-members", "5", "now"}, ExitUsage, `chama split: unexpected argument "now"`},
		{"can't be split", []string{"split", "-total", "3", "-members", "5"}, ExitError, "chama split: Ksh3 can't be split between 5 members"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := run(t, tt.args...)
			if code != tt.wantCode {
				t.Errorf("Run(%q) exit code = %d, want %d", tt.args, code, tt.wantCode)
			}
			if !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("Run(%q) stderr\n got: %q\nwant it to contain: %q", tt.args, stderr, tt.wantStderr)
			}
			if stdout != "" {
				t.Errorf("Run(%q) printed %q on stdout; problems go to stderr only", tt.args, stdout)
			}
		})
	}
}

func TestRunHelpFlag(t *testing.T) {
	for _, cmd := range []string{"split", "loan"} {
		code, stdout, stderr := run(t, cmd, "-h")
		if code != ExitOK {
			t.Errorf("Run(%q, \"-h\") exit code = %d, want 0", cmd, code)
		}
		if stdout != "" || !strings.Contains(stderr, "Usage of "+cmd) {
			t.Errorf("Run(%q, \"-h\") should print the flags to stderr\n got: stdout %q, stderr %q", cmd, stdout, stderr)
		}
	}
	_, _, stderr := run(t, "loan", "-h")
	if !strings.Contains(stderr, "(default 10)") || !strings.Contains(stderr, "-months") {
		t.Errorf("loan -h should list -months and show -rate's default of 10; got:\n%s", stderr)
	}
}
