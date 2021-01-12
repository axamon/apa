package main

import "time"

// Minuta da confermare
type Minuta struct {
	ID        string
	Testo     string
	AP        []ActionPoint
	Approvato bool
}

// ActionPoint da completare
type ActionPoint struct {
	ID          string
	Accountable string
	Responsible string
	Cosa        string
	EntroQuando time.Time
	OreUomo     uint
	Budget      uint
	Approvato   bool
}
