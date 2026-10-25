// Package pesabudget is the Phase 2 capstone: a command-line budget tracker that reads
// M-Pesa statement CSVs, sorts the spending into categories, compares it with budgets
// and prints a report. It's configured with env variables and flags.
package pesabudget

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Rule puts a transaction in Category when Keyword appears in its Details, ignoring
// case.
type Rule struct {
	Keyword  string
	Category string
}

// DefaultRules are tried after any rules from PESA_RULES. Order matters: the first
// rule that matches wins, which is why "Charge" comes first.
var DefaultRules = []Rule{
	{"Charge", "fees"},
	{"SOKO BORA", "food"},
	{"MAMA PIMA", "food"},
	{"NYAMA CHOMA", "food"},
	{"STIMA", "bills"},
	{"MAJI", "bills"},
	{"SACCO", "transport"},
	{"Airtime", "airtime"},
	{"Customer Transfer", "sent"},
}

// Other is the category for a transaction no rule matches.
const Other = "other"

// Categorise returns the Category of the first rule whose Keyword appears in details,
// ignoring case, or Other.
//
//	Categorise("Merchant Payment to 512345 - SOKO BORA SUPERMARKET", DefaultRules) → "food"
func Categorise(details string, rules []Rule) string {
	lower := strings.ToLower(details)
	for _, r := range rules {
		if strings.Contains(lower, strings.ToLower(r.Keyword)) {
			return r.Category
		}
	}
	return Other
}

// ParseRules reads rules in the form PESA_RULES uses: KEYWORD=category pairs separated
// by ";", with spaces around each part trimmed. The category is lower-cased. "" gives
// no rules. A pair with no "=" or an empty side is an error.
//
//	ParseRules("JAVA HOUSE=eating-out; Netflix=fun") → [{JAVA HOUSE eating-out} {Netflix fun}]
func ParseRules(s string) ([]Rule, error) {
	rules := []Rule{}
	if strings.TrimSpace(s) == "" {
		return rules, nil
	}
	for _, part := range strings.Split(s, ";") {
		kw, cat, ok := strings.Cut(part, "=")
		kw, cat = strings.TrimSpace(kw), strings.ToLower(strings.TrimSpace(cat))
		if !ok || kw == "" || cat == "" {
			return nil, fmt.Errorf("bad rule %q; want KEYWORD=category", strings.TrimSpace(part))
		}
		rules = append(rules, Rule{Keyword: kw, Category: cat})
	}
	return rules, nil
}

// ParseBudgets reads budgets in the form PESA_BUDGETS and -budgets use:
// category=amount pairs separated by ",", amounts in whole shillings, spaces trimmed,
// categories lower-cased. It returns them in cents. "" gives an empty map.
//
//	ParseBudgets("food=8000, transport=3500") → map[food:800000 transport:350000]
//
// Errors:
//
//	bad budget "food"; want category=amount
//	budget for "food" must be a whole number of shillings, got "8k"
//	budget for "food" given twice
func ParseBudgets(s string) (map[string]int, error) {
	budgets := map[string]int{}
	if strings.TrimSpace(s) == "" {
		return budgets, nil
	}
	for _, part := range strings.Split(s, ",") {
		cat, amount, ok := strings.Cut(part, "=")
		cat, amount = strings.ToLower(strings.TrimSpace(cat)), strings.TrimSpace(amount)
		if !ok || cat == "" || amount == "" {
			return nil, fmt.Errorf("bad budget %q; want category=amount", strings.TrimSpace(part))
		}
		if !digits(amount) || len(amount) > 12 {
			return nil, fmt.Errorf("budget for %q must be a whole number of shillings, got %q", cat, amount)
		}
		if _, dup := budgets[cat]; dup {
			return nil, fmt.Errorf("budget for %q given twice", cat)
		}
		n, _ := strconv.Atoi(amount)
		budgets[cat] = n * 100
	}
	return budgets, nil
}

// Line is one category in the report. Amounts are in cents.
type Line struct {
	Category  string
	Spent     int
	Budget    int
	HasBudget bool
}

// Summary is everything the report shows.
type Summary struct {
	Month  string // "2026-04", or "" for every month
	Lines  []Line
	PaidIn int // cents
}

// Summarise adds up the completed transactions in month ("2026-04", or "" for all) by
// category. Only withdrawals count as spending; PaidIn adds up the money received.
// A receipt number seen more than once (statements that overlap) counts only the
// first time. Every category with a budget gets a Line, even with nothing spent.
// Lines are sorted by Spent, highest first, then by Category.
func Summarise(txs []Transaction, rules []Rule, budgets map[string]int, month string) Summary {
	s := Summary{Month: month}
	spent := map[string]int{}
	seen := map[string]bool{}
	for _, tx := range txs {
		if seen[tx.Receipt] {
			continue
		}
		seen[tx.Receipt] = true
		if tx.Status != "Completed" || (month != "" && tx.Time.Format("2006-01") != month) {
			continue
		}
		s.PaidIn += tx.PaidIn
		if tx.Withdrawn > 0 {
			spent[Categorise(tx.Details, rules)] += tx.Withdrawn
		}
	}
	for cat := range budgets {
		if _, ok := spent[cat]; !ok {
			spent[cat] = 0
		}
	}
	s.Lines = []Line{}
	for cat, amount := range spent {
		b, ok := budgets[cat]
		s.Lines = append(s.Lines, Line{Category: cat, Spent: amount, Budget: b, HasBudget: ok})
	}
	sort.Slice(s.Lines, func(i, j int) bool {
		if s.Lines[i].Spent != s.Lines[j].Spent {
			return s.Lines[i].Spent > s.Lines[j].Spent
		}
		return s.Lines[i].Category < s.Lines[j].Category
	})
	return s
}

// Over reports whether any category has spent more than its budget.
func (s Summary) Over() bool {
	for _, l := range s.Lines {
		if l.HasBudget && l.Spent > l.Budget {
			return true
		}
	}
	return false
}

// Report formats a Summary as a table. Each row is fmt.Sprintf("%-12s %14s %14s %14s")
// of the category, spent, budget and left (budget minus spent), with "-" for the
// budget and left of a category that has no budget, and " OVER" after a row that is
// over its budget. The TOTAL row adds up spending, budgets, and left over the
// budgeted categories; its budget and left are "-" when there are no budgets.
//
//	Pesa Budget: 2026-04
//	CATEGORY              SPENT         BUDGET           LEFT
//	food            Ksh4,100.00    Ksh3,000.00   -Ksh1,100.00 OVER
//	bills           Ksh1,600.00              -              -
//	TOTAL           Ksh5,700.00    Ksh3,000.00   -Ksh1,100.00
//	money in: Ksh8,000.00
//
// The title says "all months" when Month is "". Every line ends with "\n".
func Report(s Summary) string {
	var b strings.Builder
	title := s.Month
	if title == "" {
		title = "all months"
	}
	fmt.Fprintf(&b, "Pesa Budget: %s\n", title)
	row := func(cat, spent, budget, left, mark string) {
		fmt.Fprintf(&b, "%-12s %14s %14s %14s%s\n", cat, spent, budget, left, mark)
	}
	row("CATEGORY", "SPENT", "BUDGET", "LEFT", "")

	totalSpent, totalBudget, totalLeft, anyBudget := 0, 0, 0, false
	for _, l := range s.Lines {
		totalSpent += l.Spent
		if !l.HasBudget {
			row(l.Category, Ksh(l.Spent), "-", "-", "")
			continue
		}
		anyBudget = true
		left := l.Budget - l.Spent
		totalBudget += l.Budget
		totalLeft += left
		mark := ""
		if left < 0 {
			mark = " OVER"
		}
		row(l.Category, Ksh(l.Spent), Ksh(l.Budget), Ksh(left), mark)
	}
	if anyBudget {
		row("TOTAL", Ksh(totalSpent), Ksh(totalBudget), Ksh(totalLeft), "")
	} else {
		row("TOTAL", Ksh(totalSpent), "-", "-", "")
	}
	fmt.Fprintf(&b, "money in: %s\n", Ksh(s.PaidIn))
	return b.String()
}

// Ksh formats cents as shillings with commas and two decimal places:
// Ksh(410000) is "Ksh4,100.00" and Ksh(-110000) is "-Ksh1,100.00".
func Ksh(cents int) string {
	sign := ""
	if cents < 0 {
		sign, cents = "-", -cents
	}
	whole := strconv.Itoa(cents / 100)
	var b strings.Builder
	for i := 0; i < len(whole); i++ {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(whole[i])
	}
	return fmt.Sprintf("%sKsh%s.%02d", sign, b.String(), cents%100)
}

// validMonth reports whether m looks like "2026-04".
func validMonth(m string) bool {
	_, err := time.Parse("2006-01", m)
	return err == nil && len(m) == 7
}
