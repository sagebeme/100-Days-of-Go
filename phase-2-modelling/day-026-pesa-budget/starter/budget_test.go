package pesabudget

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func load(t *testing.T, path string) []Transaction {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	txs, err := ParseStatement(f)
	if err != nil {
		t.Fatalf("ParseStatement(%s) returned an error: %v", path, err)
	}
	return txs
}

func TestParseStatement(t *testing.T) {
	txs := load(t, "testdata/april.csv")
	if len(txs) != 11 {
		t.Fatalf("ParseStatement(april.csv) gave %d transactions, want 11", len(txs))
	}
	tx := txs[2]
	if tx.Receipt != "SD01AAAA02" || tx.Withdrawn != 315000 || tx.Balance != 708650 || tx.PaidIn != 0 ||
		!tx.Time.Equal(time.Date(2026, 4, 1, 9, 12, 45, 0, EAT)) {
		t.Errorf("april.csv row 3 = %+v; want SD01AAAA02, Withdrawn 315000, Balance 708650, 2026-04-01 09:12:45 EAT", tx)
	}
	if got := len(load(t, "testdata/march.csv")); got != 12 {
		t.Errorf("march.csv (with a byte order mark) gave %d transactions, want 12", got)
	}

	bad := "Receipt No.,Completion Time,Details,Transaction Status,Paid In,Withdrawn,Balance\n" +
		"X1,2026-04-01 07:30:00,Airtime Purchase,Completed,,-1o0.00,100.00\n"
	if _, err := ParseStatement(strings.NewReader(bad)); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("ParseStatement with a bad amount on line 2 = %v, want an error mentioning line 2", err)
	}
}

func TestCategorise(t *testing.T) {
	tests := []struct {
		details, want string
	}{
		{"Merchant Payment to 512345 - SOKO BORA SUPERMARKET", "food"},
		{"Merchant Payment to 512345 - Soko Bora Supermarket", "food"},
		{"Pay Bill to 400200 - STIMA TOKENS Acc. 04512345678", "bills"},
		{"Pay Bill Charge", "fees"},
		{"Customer Transfer of Funds Charge", "fees"}, // "Charge" comes before "Customer Transfer"
		{"Customer Transfer to 0712***678 - WANJIKU KAMAU", "sent"},
		{"Merchant Payment to 700300 - MSAFIRI SACCO", "transport"},
		{"Airtime Purchase", "airtime"},
		{"Merchant Payment to 999999 - JAVA HOUSE", "other"},
	}
	for _, tt := range tests {
		if got := Categorise(tt.details, DefaultRules); got != tt.want {
			t.Errorf("Categorise(%q, DefaultRules) = %q, want %q", tt.details, got, tt.want)
		}
	}
	mine := []Rule{{"soko bora", "groceries"}}
	if got := Categorise("Merchant Payment to 512345 - SOKO BORA SUPERMARKET", append(mine, DefaultRules...)); got != "groceries" {
		t.Errorf("an earlier rule should win; got %q, want %q", got, "groceries")
	}
}

func TestParseRules(t *testing.T) {
	got, err := ParseRules(" JAVA HOUSE = Eating-Out ;Netflix=fun")
	want := []Rule{{"JAVA HOUSE", "eating-out"}, {"Netflix", "fun"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ParseRules\n got: %+v, %v\nwant: %+v, <nil>", got, err, want)
	}
	if got, err := ParseRules(""); err != nil || got == nil || len(got) != 0 {
		t.Errorf("ParseRules(\"\") = %#v, %v; want []Rule{}, nil", got, err)
	}
	for _, bad := range []string{"JAVA HOUSE", "=fun", "Netflix=", "a=b;;c=d"} {
		if got, err := ParseRules(bad); err == nil {
			t.Errorf("ParseRules(%q) = %+v, nil; want an error", bad, got)
		}
	}
}

func TestParseBudgets(t *testing.T) {
	got, err := ParseBudgets("food=8000, Transport = 3500,fun=0")
	want := map[string]int{"food": 800000, "transport": 350000, "fun": 0}
	if err != nil || !maps.Equal(got, want) {
		t.Errorf("ParseBudgets\n got: %v, %v\nwant: %v, <nil>", got, err, want)
	}
	if got, err := ParseBudgets(""); err != nil || got == nil || len(got) != 0 {
		t.Errorf("ParseBudgets(\"\") = %#v, %v; want an empty map, nil", got, err)
	}
	tests := []struct {
		input, want string
	}{
		{"food", `bad budget "food"; want category=amount`},
		{"food=8000,=10", `bad budget "=10"; want category=amount`},
		{"food=", `bad budget "food="; want category=amount`},
		{"food=8k", `budget for "food" must be a whole number of shillings, got "8k"`},
		{"food=-5", `budget for "food" must be a whole number of shillings, got "-5"`},
		{"food=8000.50", `budget for "food" must be a whole number of shillings, got "8000.50"`},
		{"food=1,FOOD=2", `budget for "food" given twice`},
	}
	for _, tt := range tests {
		_, err := ParseBudgets(tt.input)
		if err == nil || err.Error() != tt.want {
			t.Errorf("ParseBudgets(%q)\n got: %v\nwant: %s", tt.input, err, tt.want)
		}
	}
}

func TestKsh(t *testing.T) {
	tests := []struct {
		cents int
		want  string
	}{
		{0, "Ksh0.00"},
		{5, "Ksh0.05"},
		{410000, "Ksh4,100.00"},
		{-110000, "-Ksh1,100.00"},
		{123456789, "Ksh1,234,567.89"},
	}
	for _, tt := range tests {
		if got := Ksh(tt.cents); got != tt.want {
			t.Errorf("Ksh(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func both(t *testing.T) []Transaction {
	t.Helper()
	return append(load(t, "testdata/march.csv"), load(t, "testdata/april.csv")...)
}

func TestSummarise(t *testing.T) {
	budgets := map[string]int{"food": 300000, "transport": 50000, "fun": 100000}
	got := Summarise(both(t), DefaultRules, budgets, "2026-04")
	want := Summary{
		Month: "2026-04",
		Lines: []Line{
			{"food", 410000, 300000, true},
			{"bills", 160000, 0, false},
			{"airtime", 25000, 0, false},
			{"transport", 24000, 50000, true},
			{"fees", 1000, 0, false},
			{"fun", 0, 100000, true},
		},
		PaidIn: 800000,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Summarise(April)\n got: %+v\nwant: %+v", got, want)
	}
	if !got.Over() {
		t.Errorf("Over() = false; food is over budget")
	}
}

func TestSummariseAllMonths(t *testing.T) {
	// SC77GHIJ71 is in both files: it must count once.
	got := Summarise(both(t), DefaultRules, nil, "")
	want := Summary{
		Lines: []Line{
			{"food", 632550, 0, false},
			{"bills", 210000, 0, false},
			{"sent", 80000, 0, false},
			{"transport", 36000, 0, false},
			{"airtime", 35000, 0, false},
			{"fees", 2800, 0, false},
		},
		PaidIn: 1375000,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Summarise(all months)\n got: %+v\nwant: %+v", got, want)
	}
	if got.Over() {
		t.Errorf("Over() = true with no budgets")
	}

	empty := Summarise(nil, DefaultRules, nil, "2026-05")
	if empty.Lines == nil || len(empty.Lines) != 0 || empty.PaidIn != 0 {
		t.Errorf("Summarise(nil) = %#v, want no lines (an empty, non-nil slice) and nothing paid in", empty)
	}
}

const aprilReport = `Pesa Budget: 2026-04
CATEGORY              SPENT         BUDGET           LEFT
food            Ksh4,100.00    Ksh3,000.00   -Ksh1,100.00 OVER
bills           Ksh1,600.00              -              -
airtime           Ksh250.00              -              -
transport         Ksh240.00      Ksh500.00      Ksh260.00
fees               Ksh10.00              -              -
fun                 Ksh0.00    Ksh1,000.00    Ksh1,000.00
TOTAL           Ksh6,200.00    Ksh4,500.00      Ksh160.00
money in: Ksh8,000.00
`

const allReport = `Pesa Budget: all months
CATEGORY              SPENT         BUDGET           LEFT
food            Ksh6,325.50              -              -
bills           Ksh2,100.00              -              -
sent              Ksh800.00              -              -
transport         Ksh360.00              -              -
airtime           Ksh350.00              -              -
fees               Ksh28.00              -              -
TOTAL           Ksh9,963.50              -              -
money in: Ksh13,750.00
`

func TestReport(t *testing.T) {
	budgets := map[string]int{"food": 300000, "transport": 50000, "fun": 100000}
	if got := Report(Summarise(both(t), DefaultRules, budgets, "2026-04")); got != aprilReport {
		t.Errorf("Report(April)\n got:\n%s\nwant:\n%s", got, aprilReport)
	}
	if got := Report(Summarise(both(t), DefaultRules, nil, "")); got != allReport {
		t.Errorf("Report(all months)\n got:\n%s\nwant:\n%s", got, allReport)
	}
}

// fakeEnv is a LookupFunc backed by a map, so the tests never read your real
// environment.
func fakeEnv(m map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func run(t *testing.T, env map[string]string, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(args, fakeEnv(env), &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunReports(t *testing.T) {
	files := []string{"testdata/march.csv", "testdata/april.csv"}
	tests := []struct {
		name string
		env  map[string]string
		args []string
		code int
		want string
	}{
		{"flags", nil, append([]string{"-month", "2026-04", "-budgets", "food=3000,transport=500,fun=1000"}, files...), ExitOK, aprilReport},
		{"env", map[string]string{"PESA_MONTH": "2026-04", "PESA_BUDGETS": "food=3000,transport=500,fun=1000"}, files, ExitOK, aprilReport},
		{"flags beat env", map[string]string{"PESA_MONTH": "2026-03", "PESA_BUDGETS": "food=1"}, append([]string{"-month=2026-04", "-budgets=food=3000,transport=500,fun=1000"}, files...), ExitOK, aprilReport},
		{"empty env means unset", map[string]string{"PESA_MONTH": "", "PESA_BUDGETS": "", "PESA_STRICT": ""}, files, ExitOK, allReport},
		{"strict and over", nil, append([]string{"-month", "2026-04", "-strict", "-budgets", "food=3000,transport=500,fun=1000"}, files...), ExitOverBudget, aprilReport},
		{"strict from env", map[string]string{"PESA_STRICT": "true"}, append([]string{"-month", "2026-04", "-budgets", "food=3000,transport=500,fun=1000"}, files...), ExitOverBudget, aprilReport},
		{"-strict=false beats env", map[string]string{"PESA_STRICT": "true"}, append([]string{"-strict=false", "-month", "2026-04", "-budgets", "food=3000,transport=500,fun=1000"}, files...), ExitOK, aprilReport},
		{"strict but within budget", nil, append([]string{"-strict"}, files...), ExitOK, allReport},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := run(t, tt.env, tt.args...)
			if code != tt.code || stdout != tt.want || stderr != "" {
				t.Errorf("Run(%q)\n got: exit %d, stderr %q, stdout:\n%s\nwant: exit %d, no stderr, stdout:\n%s", tt.args, code, stderr, stdout, tt.code, tt.want)
			}
		})
	}
}

func TestRunRulesFromEnv(t *testing.T) {
	env := map[string]string{"PESA_RULES": "SOKO BORA=groceries;NYAMA CHOMA=eating-out"}
	code, stdout, stderr := run(t, env, "-month", "2026-04", "testdata/april.csv")
	if code != ExitOK || stderr != "" {
		t.Fatalf("Run with PESA_RULES: exit %d, stderr %q", code, stderr)
	}
	for _, want := range []string{"groceries       Ksh3,150.00", "eating-out        Ksh950.00"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("Run with PESA_RULES: the report doesn't contain %q:\n%s", want, stdout)
		}
	}
}

func TestRunProblems(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.csv")
	csv := "Receipt No.,Completion Time,Details,Transaction Status,Paid In,Withdrawn,Balance\n" +
		"X1,2026-04-01 07:30:00,Airtime Purchase,Completed,,-100.00,100.00\n" +
		"X2,yesterday,Airtime Purchase,Completed,,-100.00,0.00\n"
	if err := os.WriteFile(broken, []byte(csv), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "missing.csv")

	tests := []struct {
		name string
		env  map[string]string
		args []string
		code int
		want string // stderr must contain this
	}{
		{"no files", nil, nil, ExitUsage, "pesabudget: give at least one statement CSV file"},
		{"unknown flag", nil, []string{"-year", "2026", "testdata/april.csv"}, ExitUsage, "flag provided but not defined: -year"},
		{"bad month flag", nil, []string{"-month", "2026-4", "testdata/april.csv"}, ExitUsage, `pesabudget: month "2026-4" must look like 2026-04`},
		{"bad month env", map[string]string{"PESA_MONTH": "April"}, []string{"testdata/april.csv"}, ExitUsage, `pesabudget: month "April" must look like 2026-04`},
		{"bad budgets flag", nil, []string{"-budgets", "food=8k", "testdata/april.csv"}, ExitUsage, `pesabudget: -budgets: budget for "food" must be a whole number of shillings, got "8k"`},
		{"bad budgets env", map[string]string{"PESA_BUDGETS": "food"}, []string{"testdata/april.csv"}, ExitUsage, `pesabudget: PESA_BUDGETS: bad budget "food"; want category=amount`},
		{"bad strict env", map[string]string{"PESA_STRICT": "yes"}, []string{"testdata/april.csv"}, ExitUsage, `pesabudget: PESA_STRICT must be true or false, got "yes"`},
		{"bad rules env", map[string]string{"PESA_RULES": "JAVA HOUSE"}, []string{"testdata/april.csv"}, ExitUsage, `pesabudget: PESA_RULES: bad rule "JAVA HOUSE"`},
		{"missing file", nil, []string{"testdata/april.csv", missing}, ExitError, "pesabudget: " + missing + ":"},
		{"broken file", nil, []string{broken}, ExitError, "pesabudget: " + broken + ": line 3:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := run(t, tt.env, tt.args...)
			if code != tt.code {
				t.Errorf("Run(%q) exit code = %d, want %d", tt.args, code, tt.code)
			}
			if !strings.Contains(stderr, tt.want) {
				t.Errorf("Run(%q) stderr\n got: %q\nwant it to contain: %q", tt.args, stderr, tt.want)
			}
			if stdout != "" {
				t.Errorf("Run(%q) printed a report although it failed:\n%s", tt.args, stdout)
			}
		})
	}
}
