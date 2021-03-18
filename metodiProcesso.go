package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Metodi

func (p Processo) QuantiInput() int {
	return len(p.Input)
}

func (p Processo) QuantiOutput() int {
	return len(p.Output)
}

// UOCoinvolte restituisce la lista delle Unità organizzative
// coinvolte in un processo.
func (p Processo) UOCoinvolte() []string {
	var uos []string
	for _, id := range p.Attivitas {
		var att Attivita
		a.Find(bson.M{"id": id}).One(&att)
		uos = append(uos, att.UO)
		fmt.Println(a)
	}
	return uos
}

func (p *Processo) HaAMonte(p2 *Processo) {
	// aggiorna processo a monte
	if !find(p2.Output, p.Id) { // se non è già presente
		p2.Output = append(p2.Output, p.Id)
		p2.Update()
	}
	// aggiorna processo a valle
	if !find(p.Input, p2.Id) {
		p.Input = append(p.Input, p2.Id)
		p.Update()
	}
}

func (p Processo) HaAValle(p2 Processo) {
	if !find(p.Output, p2.Id) {
		p.Output = append(p.Output, p2.Id)
		p.Update()
	}
	if !find(p2.Input, p.Id) {
		p2.Input = append(p2.Input, p.Id)
		p2.Update()
	}
}

func (p *Processo) Approva() {
	p.Status = Approvato
	p.Update()
}

func (p Processo) Ver() uint {
	return p.Versione
}

func (p Processo) Delete() {
	err := c.Remove(bson.M{"id": p.Id})
	if err != nil {
		log.Printf("Delete di %s in errore: %v\n", p.Titolo, err)
	}
}

func (p *Processo) Update() error {
	p.Versione++
	p.Updated_at = time.Now()

	err := c.Update(bson.M{"id": p.Id}, &p)
	if err != nil {
		log.Printf("Update di %s in errore: %v\n", p.Id, err)
	}
	return err
}

func (att Attivita) Save() error {
	err := a.Insert(&att)
	if err != nil {
		log.Printf("Save di %s in errore: %v\n", att.Id, err)
	}
	return err
}

func (p Processo) Save() error {
	err := c.Insert(&p)
	if err != nil {
		log.Printf("Save di %s in errore: %v\n", p.Id, err)

	}
	return err
}

func find(s []string, val string) bool {
	for _, item := range s {
		if item == val {
			return true
		}
	}
	return false
}
