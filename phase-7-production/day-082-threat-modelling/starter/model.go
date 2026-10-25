// Package threatmodelling describes a system as components and data flows, lists the
// STRIDE threats to ask about for each one, and ranks risks by likelihood × impact.
package threatmodelling

import (
	"errors"
)

// Kind is what sort of thing a component is in a data-flow diagram.
type Kind string

// The three kinds of box in a data-flow diagram.
const (
	External  Kind = "external"  // a person or system you don't run: a passenger, M-Pesa
	Process   Kind = "process"   // code you run: the booking API
	DataStore Kind = "datastore" // where data rests: Postgres, Redis, a bucket
)

// Component is one box in the diagram. Zone names its trust zone, such as "internet"
// or "internal"; a flow between different zones crosses a trust boundary.
type Component struct {
	Name string
	Kind Kind
	Zone string
}

// Flow is one arrow: data moving from one component to another.
type Flow struct {
	From, To      string
	Data          string // what moves, such as "booking request"
	Encrypted     bool   // TLS or similar on the wire
	Authenticated bool   // the receiver checks who the sender is
}

// Model is the whole diagram.
type Model struct {
	Components []Component
	Flows      []Flow
}

// Category is one letter of STRIDE.
type Category string

// The six STRIDE categories, in order.
const (
	Spoofing              Category = "Spoofing"
	Tampering             Category = "Tampering"
	Repudiation           Category = "Repudiation"
	InformationDisclosure Category = "Information disclosure"
	DenialOfService       Category = "Denial of service"
	ElevationOfPrivilege  Category = "Elevation of privilege"
)

// Order lists the categories in STRIDE order.
var Order = []Category{Spoofing, Tampering, Repudiation, InformationDisclosure, DenialOfService, ElevationOfPrivilege}

// PerKind says which categories apply to each kind of component (the classic
// "STRIDE per element" table).
var PerKind = map[Kind][]Category{
	External:  {Spoofing, Repudiation},
	Process:   {Spoofing, Tampering, Repudiation, InformationDisclosure, DenialOfService, ElevationOfPrivilege},
	DataStore: {Tampering, InformationDisclosure, DenialOfService},
}

// PerFlow lists the categories that apply to every data flow. A flow that crosses a
// trust boundary without authentication also gets Spoofing: anyone could send it.
var PerFlow = []Category{Tampering, InformationDisclosure, DenialOfService}

// Questions holds the question to ask for each category; %s is the target's name.
var Questions = map[Category]string{
	Spoofing:              "Could someone pretend to be %s?",
	Tampering:             "Could someone change %s or the data in it?",
	Repudiation:           "Could %s deny doing something, and could we prove otherwise?",
	InformationDisclosure: "Could %s leak data to someone who shouldn't see it?",
	DenialOfService:       "Could someone stop %s from working?",
	ElevationOfPrivilege:  "Could someone use %s to do more than they're allowed?",
}

// Threat is one question to answer about one target.
type Threat struct {
	Category Category
	Target   string // a component name, or "From -> To" for a flow
	Question string
	Boundary bool // the target is a flow that crosses a trust boundary
}

// Validation errors.
var (
	ErrDuplicate = errors.New("duplicate component")
	ErrUnknown   = errors.New("flow names an unknown component")
	ErrBadKind   = errors.New("unknown component kind")
)

// Validate checks that component names are unique, every kind is one of the three,
// and every flow's From and To name a component. Errors wrap the ones above and name
// the culprit.
func (m Model) Validate() error {
	// Your code here
	return nil
}

// CrossesBoundary reports whether f joins components in different trust zones.
// The model must be valid.
func (m Model) CrossesBoundary(f Flow) bool {
	// Your code here
	return false
}

func (m Model) zone(name string) string {
	// Your code here
	return ""
}

// STRIDE validates the model and returns its threats: first each component in model
// order with the categories PerKind gives its kind, then each flow in model order with
// PerFlow, plus Spoofing first for a flow that crosses a boundary unauthenticated.
// Within each target, categories come in STRIDE order. Each question is
// fmt.Sprintf(Questions[category], target).
func STRIDE(m Model) ([]Threat, error) {
	// Your code here
	return nil, nil
}

// ErrScale is returned for a likelihood or impact outside 1 to 5.
var ErrScale = errors.New("likelihood and impact must be 1 to 5")

// Score returns likelihood × impact, each on a scale of 1 (rare, minor) to 5 (almost
// certain, severe), or ErrScale.
//
//	Score(4, 5) → 20, nil
func Score(likelihood, impact int) (int, error) {
	// Your code here
	return 0, nil
}

// Level names a score: 1 to 4 "low", 5 to 9 "medium", 10 to 16 "high", 17 and up
// "critical".
func Level(score int) string {
	// Your code here
	return ""
}

// Risk is a threat someone has rated.
type Risk struct {
	Threat     Threat
	Likelihood int
	Impact     int
}

// Prioritise returns a sorted copy of risks, most urgent first: higher score, then
// higher impact (a rare disaster beats a frequent nuisance), then STRIDE order, then
// target name. It returns ErrScale, wrapped with the target, for any bad rating.
func Prioritise(risks []Risk) ([]Risk, error) {
	// Your code here
	return nil, nil
}
