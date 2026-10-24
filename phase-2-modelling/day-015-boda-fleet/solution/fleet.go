// Package bodafleet keeps track of a small fleet of boda bodas: who rides each bike,
// how far it has gone, what it has earned and when it needs a service.
package bodafleet

import (
	"errors"
	"fmt"
	"sort"
)

// ServiceInterval is how many kilometres a bike can go between services.
const ServiceInterval = 1000

// Bike is one boda boda in the fleet.
type Bike struct {
	Plate         string // number plate, e.g. "KMFA 123B"
	Rider         string // who rides it today; "" if nobody
	Odometer      int    // kilometres on the clock
	LastServiceKm int    // the odometer reading at the last service
	Trips         int    // trips ridden
	Earnings      int    // fares collected, in Ksh
}

// Label names the bike and its rider for a list:
//
//	Bike{Plate: "KMFA 123B", Rider: "Otieno"}.Label() → "KMFA 123B (Otieno)"
//	Bike{Plate: "KMGG 404Z"}.Label()                  → "KMGG 404Z (no rider)"
func (b Bike) Label() string {
	rider := b.Rider
	if rider == "" {
		rider = "no rider"
	}
	return fmt.Sprintf("%s (%s)", b.Plate, rider)
}

// Ride records one trip of km kilometres that paid fare shillings. It adds km to the
// odometer, fare to the earnings and one to the trip count. A trip must be at least
// 1 km and a fare can't be negative; on an error the bike is left unchanged.
//
// The receiver is a pointer, *Bike, so Ride changes the caller's bike, not a copy.
func (b *Bike) Ride(km, fare int) error {
	if km < 1 {
		return fmt.Errorf("trip of %d km: a trip must be at least 1 km", km)
	}
	if fare < 0 {
		return fmt.Errorf("fare of Ksh%d: a fare can't be negative", fare)
	}
	b.Odometer += km
	b.Earnings += fare
	b.Trips++
	return nil
}

// NeedsService reports whether the bike has gone ServiceInterval km or more since its
// last service.
func (b Bike) NeedsService() bool {
	return b.Odometer-b.LastServiceKm >= ServiceInterval
}

// Service marks the bike as serviced at its current odometer reading.
func (b *Bike) Service() {
	b.LastServiceKm = b.Odometer
}

// Fleet is every bike one owner runs.
type Fleet struct {
	Bikes []Bike
}

// Add puts a bike in the fleet. It returns an error if the plate is empty or a bike
// with the same plate is already there.
func (f *Fleet) Add(b Bike) error {
	if b.Plate == "" {
		return errors.New("a bike needs a plate")
	}
	if f.Find(b.Plate) != nil {
		return fmt.Errorf("%s is already in the fleet", b.Plate)
	}
	f.Bikes = append(f.Bikes, b)
	return nil
}

// Find returns a pointer to the bike with this plate, so the caller can change it, or
// nil if there's no such bike.
func (f *Fleet) Find(plate string) *Bike {
	for i := range f.Bikes {
		if f.Bikes[i].Plate == plate {
			return &f.Bikes[i]
		}
	}
	return nil
}

// Ride records a trip on the bike with this plate. It returns an error if there's no
// such bike or the trip itself is invalid.
func (f *Fleet) Ride(plate string, km, fare int) error {
	b := f.Find(plate)
	if b == nil {
		return fmt.Errorf("no bike with plate %s", plate)
	}
	return b.Ride(km, fare)
}

// TotalEarnings adds up what every bike in the fleet has earned.
func (f Fleet) TotalEarnings() int {
	total := 0
	for _, b := range f.Bikes {
		total += b.Earnings
	}
	return total
}

// DueForService returns the plates of the bikes that need a service, in alphabetical
// order. It returns an empty (non-nil) slice when none do.
func (f Fleet) DueForService() []string {
	due := []string{}
	for _, b := range f.Bikes {
		if b.NeedsService() {
			due = append(due, b.Plate)
		}
	}
	sort.Strings(due)
	return due
}
