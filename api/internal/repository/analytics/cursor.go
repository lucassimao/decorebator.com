package analytics

import "time"

// WordMasteryCursor preserves the established mastery-first order. Mastery is
// kept as its canonical NUMERIC text representation so a page boundary does
// not lose precision by round-tripping through float64.
type WordMasteryCursor struct {
	MasteryLevel string
	LastSeenAt   *time.Time
	WordID       int64
}
