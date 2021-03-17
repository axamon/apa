package main

import (
	"time"
)

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

//TemplateData holds data to pass to templates.
type TemplateData struct {
	PageTitle string
	ID        string
}

type StatusType string

const (
	Nuovo      StatusType = "Nuovo"
	Verificato StatusType = "Verificato"
	Approvato  StatusType = "Approvato"
	InVigore   StatusType = "In Vigore"
	Superato   StatusType = "Superato"
)

type RaciResp string

const (
	R RaciResp = "R"
	A RaciResp = "A"
	C RaciResp = "C"
	I RaciResp = "I"
)

type Processo struct {
	Id             string      `json:"id"`
	Titolo         string      `json:"titolo,omitempty"`
	Descrizione    string      `json:"descrizione,omitempty"`
	Testo          string      `json:"testo,omitempty"`
	Autori         []string    `json:"autori,omitempty"`
	Verificatori   []string    `json:"verificatori,omitempty"`
	Approvatori    []string    `json:"approvatori,omitempty"`
	Versione       uint        `json:"versione,omitempty"`
	Input          []string    `json:"input,omitempty"`
	Output         []string    `json:"output,omitempty"`
	Attivitas      []*Attivita `json:"attivitas,omitempty"`
	Status         StatusType  `json:"status,omitempty"`
	Kpis           []Kpi       `json:"kpis,omitempty"`
	Created_at     time.Time   `json:"createdat,omitempty"`
	Updated_at     time.Time   `json:"updatedat,omitempty"`
	FlussoImmagine string      `json:"flussoimmagine,omitempty"`
	Costo          float64     `json:"costo,omitempty"`
	Tmedio         float64     `json:"tmedio,omitempty"`
	Devstd         float64     `json:"devstd,omitempty"`
}

type Attivita struct {
	Id          string
	Num         int
	UO          string
	Titolo      string
	Descrizione string
	Ruolo       RaciResp
	Input       []string
	Output      []string
	Tmedio      float64
	Devstd      float64
	Costo       float64
}

type Kpi struct {
}

type meta struct {
}
