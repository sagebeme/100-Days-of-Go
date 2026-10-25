package csvstatements

import (
	"maps"
	"os"
	"strings"
	"testing"
	"time"
)

func parseFile(t *testing.T, path string) []Transaction {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	txs, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse(%s) returned an error: %v", path, err)
	}
	return txs
}

func TestParseStatement(t *testing.T) {
	txs := parseFile(t, "testdata/statement.csv")
	if len(txs) != 12 {
		t.Fatalf("Parse(testdata/statement.csv) gave %d transactions, want 12", len(txs))
	}

	first := Transaction{
		Receipt:   "SC71ABCD12",
		Time:      time.Date(2026, 3, 1, 8, 15, 42, 0, EAT),
		Details:   "Funds received from 0722***111 - JOSEPH OTIENO",
		Status:    "Completed",
		PaidIn:    500000,
		Withdrawn: 0,
		Balance:   525000,
	}
	if got := txs[0]; !got.Time.Equal(first.Time) || got.Receipt != first.Receipt || got.Details != first.Details ||
		got.Status != first.Status || got.PaidIn != first.PaidIn || got.Withdrawn != first.Withdrawn || got.Balance != first.Balance {
		t.Errorf("first transaction\n got: %+v\nwant: %+v", got, first)
	}
	if _, offset := txs[0].Time.Zone(); offset != 3*60*60 {
		t.Errorf("times should be in EAT (UTC+3); got offset %d seconds", offset)
	}

	second := txs[1]
	if second.Withdrawn != 123550 || second.PaidIn != 0 || second.Balance != 401450 {
		t.Errorf("second transaction: Withdrawn %d, PaidIn %d, Balance %d; want 123550, 0, 401450", second.Withdrawn, second.PaidIn, second.Balance)
	}

	quoted := txs[7]
	if want := "Merchant Payment to 600100 - MAMA PIMA GROCERS, KASARANI"; quoted.Details != want {
		t.Errorf("a quoted cell with a comma\n got: %q\nwant: %q", quoted.Details, want)
	}
	if txs[6].Status != "Failed" {
		t.Errorf("txs[6].Status = %q, want %q", txs[6].Status, "Failed")
	}
}

func TestParseColumnsInAnyOrder(t *testing.T) {
	csv := "Balance, Paid In ,Withdrawn,Details,Receipt No.,Transaction Status,Completion Time,Notes\n" +
		"\"1,000.50\",,-99.5,Airtime Purchase,SD10XYZ001,Completed,2026-04-01 09:00:00,mine\n" +
		"\n" +
		"\"11,000.50\",\"10,000\",,Funds received from 0700***000 - CHAMA,SD10XYZ002,Completed,2026-04-02 10:30:00,\n"
	txs, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("Parse gave %d transactions, want 2 (blank lines are skipped)", len(txs))
	}
	if got := txs[0]; got.Receipt != "SD10XYZ001" || got.Withdrawn != 9950 || got.Balance != 100050 || got.Details != "Airtime Purchase" {
		t.Errorf("row 1 = %+v; want Receipt SD10XYZ001, Withdrawn 9950, Balance 100050, Details \"Airtime Purchase\"", got)
	}
	if got := txs[1]; got.PaidIn != 1000000 || got.Balance != 1100050 {
		t.Errorf("row 2: PaidIn %d, Balance %d; want 1000000, 1100050", got.PaidIn, got.Balance)
	}
}

const header = "Receipt No.,Completion Time,Details,Transaction Status,Paid In,Withdrawn,Balance\n"

func TestParseBOM(t *testing.T) {
	csv := "\ufeff" + header + "SE01AAA001,2026-05-01 12:00:00,Airtime Purchase,Completed,,-50.00,950.00\n"
	txs, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse with a byte order mark returned an error: %v", err)
	}
	if len(txs) != 1 || txs[0].Receipt != "SE01AAA001" {
		t.Errorf("Parse with a byte order mark = %+v, want one row with Receipt SE01AAA001", txs)
	}
}

func TestParseEmptyStatement(t *testing.T) {
	txs, err := Parse(strings.NewReader(header))
	if err != nil || txs == nil || len(txs) != 0 {
		t.Errorf("Parse(header only) = %#v, %v; want []Transaction{}, nil", txs, err)
	}
	if _, err := Parse(strings.NewReader("")); err == nil {
		t.Errorf("Parse(\"\") = nil error, want an error")
	}
}

func TestParseErrors(t *testing.T) {
	ok := "SE01AAA001,2026-05-01 12:00:00,Airtime Purchase,Completed,,-50.00,950.00\n"
	tests := []struct {
		name string
		csv  string
		want string // the error must contain this
	}{
		{"missing column", "Receipt No.,Completion Time,Details,Transaction Status,Paid In,Balance\n", `line 1: missing column "Withdrawn"`},
		{"bad time", header + ok + "SE01AAA002,01/05/2026 12:00,Airtime Purchase,Completed,,-50.00,900.00\n", "line 3:"},
		{"bad amount", header + "SE01AAA002,2026-05-01 12:00:00,Airtime Purchase,Completed,,-fifty,900.00\n", "line 2:"},
		{"three decimals", header + ok + ok + "SE01AAA002,2026-05-01 12:00:00,Airtime Purchase,Completed,,-50.005,900.00\n", "line 4:"},
		{"negative paid in", header + "SE01AAA002,2026-05-01 12:00:00,Refund,Completed,-50.00,,900.00\n", "line 2:"},
		{"positive withdrawn", header + "SE01AAA002,2026-05-01 12:00:00,Airtime Purchase,Completed,,50.00,900.00\n", "line 2:"},
		{"bad balance", header + "SE01AAA002,2026-05-01 12:00:00,Airtime Purchase,Completed,,-50.00,lots\n", "line 2:"},
		{"too few cells", header + ok + "SE01AAA002,2026-05-01 12:00:00,Airtime Purchase\n", "line 3"},
		{"unclosed quote", header + "SE01AAA002,2026-05-01 12:00:00,\"Airtime Purchase,Completed,,-50.00,900.00\n", "line"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs, err := Parse(strings.NewReader(tt.csv))
			if err == nil {
				t.Fatalf("Parse = %+v, nil; want an error containing %q", txs, tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Parse error\n got: %v\nwant it to contain: %q", err, tt.want)
			}
		})
	}
}

func TestTotals(t *testing.T) {
	txs := parseFile(t, "testdata/statement.csv")
	in, out := Totals(txs)
	// The failed Ksh2,000 payment doesn't count.
	if in != 575000 || out != 376350 {
		t.Errorf("Totals(statement) = %d, %d; want 575000, 376350", in, out)
	}
	if in, out := Totals(nil); in != 0 || out != 0 {
		t.Errorf("Totals(nil) = %d, %d; want 0, 0", in, out)
	}
}

func TestCounterparty(t *testing.T) {
	tests := []struct {
		details, want string
	}{
		{"Merchant Payment to 512345 - SOKO BORA SUPERMARKET", "SOKO BORA SUPERMARKET"},
		{"Customer Transfer to 0712***678 - WANJIKU KAMAU", "WANJIKU KAMAU"},
		{"Pay Bill to 400200 - STIMA TOKENS Acc. 04512345678", "STIMA TOKENS Acc. 04512345678"},
		{"Merchant Payment to 600100 - MAMA PIMA GROCERS - KASARANI", "MAMA PIMA GROCERS - KASARANI"},
		{"Airtime Purchase", "Airtime Purchase"},
		{"  Pay Bill Charge ", "Pay Bill Charge"},
	}
	for _, tt := range tests {
		if got := Counterparty(tt.details); got != tt.want {
			t.Errorf("Counterparty(%q)\n got: %q\nwant: %q", tt.details, got, tt.want)
		}
	}
}

func TestSpentWith(t *testing.T) {
	got := SpentWith(parseFile(t, "testdata/statement.csv"))
	want := map[string]int{
		"SOKO BORA SUPERMARKET":             187550, // 1,235.50 + 640.00; the failed 2,000 doesn't count
		"STIMA TOKENS Acc. 04512345678":     50000,
		"Pay Bill Charge":                   500,
		"WANJIKU KAMAU":                     80000,
		"Customer Transfer of Funds Charge": 1300,
		"MAMA PIMA GROCERS, KASARANI":       35000,
		"MSAFIRI SACCO":                     12000,
		"Airtime Purchase":                  10000,
	}
	if !maps.Equal(got, want) {
		t.Errorf("SpentWith(statement)\n got: %v\nwant: %v", got, want)
	}
}
