// Package bundlecalculator works out how much data your shillings buy.
package bundlecalculator

import "math"

// BundlesYouCanBuy returns how many whole bundles costing price shillings fit in budget.
// price is always more than 0.
//
//	BundlesYouCanBuy(250, 99) → 2
func BundlesYouCanBuy(budget, price int) int {
	return budget / price
}

// Change returns the shillings left after buying as many bundles as budget allows.
//
//	Change(250, 99) → 52
func Change(budget, price int) int {
	return budget % price
}

// GBToMB converts gigabytes to megabytes, dropping any part of a megabyte.
// One gigabyte is 1024 megabytes.
//
//	GBToMB(1.5) → 1536
//	GBToMB(0.3) → 307
func GBToMB(gb float64) int {
	return int(gb * 1024)
}

// PricePerGB returns what one gigabyte costs in a bundle of mb megabytes sold for
// price shillings, rounded to 2 decimal places.
//
//	PricePerGB(99, 1024) → 99
//	PricePerGB(50, 350)  → 146.29
func PricePerGB(price, mb int) float64 {
	gb := float64(mb) / 1024
	perGB := float64(price) / gb
	return math.Round(perGB*100) / 100
}
