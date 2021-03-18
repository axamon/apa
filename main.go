package main

import (
	"context"
	"crypto/md5"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"

	"golang.org/x/net/http2"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

// User holds a users account information
type User struct {
	Username      string
	Authenticated bool
}

// store will hold all session data
var store *sessions.CookieStore

// tpl holds all parsed templates
var tpl *template.Template

func init() {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	store = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	store.Options = &sessions.Options{
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(User{})

	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

var i chan string

var ctx context.Context

var c, a *mgo.Collection
var database = "db"
var mongoURL = "mongodb://localhost"

func main() {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Print("session", err)
	}

	c = session.DB(database).C("processi")
	a = session.DB(database).C("attivita")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			fmt.Println("ricevuto ctrl+c")
			os.Exit(0)
		case <-ctx.Done():
		}
	}()

	err = deleteAllProcessi()
	if err != nil {
		log.Println(err)
	}

	_, err = NewProcesso("Ciclo Passivo2")
	if err != nil {
		log.Println(err)
	}

	p, err := NewProcesso("Ciclo Passivo")
	if err != nil {
		log.Println(err)
	}

	p.Testo = "come spendiampo il prezioso budget di mamma TIM per soddisfare tuute le esigenze delle varie funzioni presenti in CTIO lorem libsum docet cane nero ibis redibis non morieris in bello"
	p.Autori = append(p.Autori, "Alberto Bregliano")
	p.Verificatori = append(p.Verificatori, "Valeria Allegretti")
	p.Approvatori = append(p.Approvatori, "Guido Bruno")

	attivita1 := Attivita{
		Id:     uuid.NewString(),
		Num:    1,
		Titolo: "Spendere",
		UO:     "CTIO.5GDT.PDT",
		Ruolo:  R,
	}

	attivita1.Save()

	fmt.Println()
	fmt.Println(attivita1)

	attivita2 := Attivita{
		Id:     uuid.NewString(),
		Num:    2,
		Titolo: "Spendere tutto",
		UO:     "CTIO.5GDT.PDO",
		Ruolo:  R,
	}

	attivita2.Save()

	p.Attivitas = append(p.Attivitas, attivita1.Id, attivita2.Id)
	p.Approva()
	err = p.Update()
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("%+v\n", p)

	VerificaBudget, _ := NewProcesso("Verifica Budget")

	attivita1a := Attivita{
		UO:          "CFO",
		Titolo:      "Verifica mensile",
		Descrizione: "ogni mese controllo spese",
		Id:          uuid.NewString(),
		Num:         1,
		Ruolo:       R,
	}
	attivita1a.Save()

	attivita2a := Attivita{
		UO:          "CTIO.5GDT.PDT",
		Titolo:      "Consegna spese",
		Descrizione: "consegna scontrini delle merendine",
		Id:          uuid.NewString(),
		Num:         2,
		Ruolo:       R,
	}

	attivita2a.Save()

	VerificaBudget.Attivitas = append(VerificaBudget.Attivitas, attivita1a.Id, attivita2a.Id)
	VerificaBudget.HaAMonte(&p)

	VerificaBudget.Update()

	var addr = flag.String("addr", ":443", "Porta TCP da usare")
	flag.Parse()

	go aggiornaTemplates()

	//r := http.NewServeMux()
	r := mux.NewRouter()

	//fs := http.FileServer(http.Dir("./static"))
	//r.Handle("/static/", http.StripPrefix("/static", fs))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/", index)
	r.Handle("/favicon.ico", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/processi", all)
	r.HandleFunc("/processo/{ID}", getProcesso)

	r.HandleFunc("/attivita", allAttivita)
	r.HandleFunc("/attivita/{ID}", getAttivita)
	r.HandleFunc("/nuovoprocesso", nuovoprocesso)
	r.HandleFunc("/deleteall", deleteall)
	r.HandleFunc("/modificaprocesso", modificaprocesso)

	r.HandleFunc("/doc/{ID}", doc)

	var srv http.Server
	srv.Addr = *addr
	srv.Handler = r
	//Enable http2
	http2.ConfigureServer(&srv, nil)

	srv.ListenAndServeTLS("certs/localhost.cert", "certs/localhost.key")

	// http.ListenAndServe(":80", r)
}

func deleteall(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		deleteAllProcessi()
		deleteAllAttivita()
		fmt.Fprint(w, http.StatusAccepted)

	default:
		fmt.Fprint(w, http.StatusForbidden)
	}
}

func modificaprocesso(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PATCH":

		headerContentTtype := r.Header.Get("Content-Type")
		if headerContentTtype != "application/json" {
			fmt.Fprint(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		var p Processo
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&p)
		if err != nil {
			log.Println(err)
		}
		p.Updated_at = time.Now()

		err = c.Update(bson.M{"id": p.Id}, &p)
		if err != nil {
			fmt.Fprint(w, http.StatusInternalServerError)
			return
		}
		var r Processo
		c.Find(bson.M{"id": p.Id}).One(&r)
		fmt.Fprintf(w, "%+v", r)

	default:
		fmt.Fprint(w, http.StatusForbidden)
	}
}

func nuovoprocesso(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

		headerContentTtype := r.Header.Get("Content-Type")
		if headerContentTtype != "application/json" {
			fmt.Fprint(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		var p Processo
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&p)
		if err != nil {
			log.Println(err)
		}

		np, err := NewProcesso(p.Titolo)
		if err != nil {
			fmt.Fprint(w, http.StatusConflict)
			return
		}
		//c.Upsert(bson.M{"id": p.Id}, &np)

		fmt.Fprintf(w, "%+v", np)

	default:
		fmt.Fprint(w, http.StatusForbidden)
	}
}

func aggiornaTemplates() {
	for {
		var err error
		tpl, err = template.ParseGlob("templates/*.gohtml")
		if err != nil {
			log.Println(err)
		}

		time.Sleep(5 * time.Second)
	}
}

func allAttivita(w http.ResponseWriter, r *http.Request) {

	attivita, err := GetAllAttivita()
	fmt.Println(attivita)
	if err != nil {
		log.Println(attivita, err)
		fmt.Fprint(w, http.StatusNotFound)
		return
	}
	bData, err := json.Marshal(attivita)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(bData))

}

func all(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	if pusher, ok := w.(http.Pusher); ok {
		// Push is supported.
		options := &http.PushOptions{
			Header: http.Header{
				"Accept-Encoding": r.Header["Accept-Encoding"],
			},
		}
		if err := pusher.Push("/static/app.ts", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}
		if err := pusher.Push("/static/typescript.min.js", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}
		if err := pusher.Push("/static/typescript.compile.min.js", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}

	}

	processi, err := GetAllProcessi()
	if err != nil {
		log.Println(err)
	}

	start := time.Now()
	tpl.ExecuteTemplate(w, "allprocessi.gohtml", processi)
	fmt.Println(time.Since(start))

}

func getProcesso(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	fmt.Println(params["ID"])

	id := params["ID"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	p, err := GetProcesso(id)
	if err != nil {
		fmt.Fprint(w, http.StatusNotFound)
		return
	}

	fmt.Fprint(w, p)

}

func getAttivita(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	fmt.Println(params["ID"])

	id := params["ID"]

	p, err := GetAttivita(id)
	if err != nil {
		fmt.Fprint(w, http.StatusNotFound)
		return
	}

	fmt.Fprint(w, p)

}

func doc(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	fmt.Println(params["ID"])

	id := params["ID"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	if pusher, ok := w.(http.Pusher); ok {
		// Push is supported.
		options := &http.PushOptions{
			Header: http.Header{
				"Accept-Encoding": r.Header["Accept-Encoding"],
			},
		}
		if err := pusher.Push("/static/app.ts", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}
		if err := pusher.Push("/static/typescript.min.js", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}
		if err := pusher.Push("/static/typescript.compile.min.js", options); err != nil {
			log.Printf("Failed to push: %v", err)
		}

	}

	p, err := GetProcesso(id)
	if err != nil {
		log.Println(err)
	}

	start := time.Now()
	tpl.ExecuteTemplate(w, "doc3.gohtml", p)
	fmt.Println(time.Since(start))

}

func nuova(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	var td TemplateData
	td.PageTitle = "Nuova Minuta"
	td.ID = <-i

	tpl.ExecuteTemplate(w, "nuova.gohtml", td)
	fmt.Println(time.Since(start))

}

func index(w http.ResponseWriter, r *http.Request) {

	data := Minuta{Testo: "ciao", ID: "1", Approvato: false, AP: []ActionPoint{
		ActionPoint{
			ID:          "101",
			Accountable: "Pasquale",
			Responsible: "Maria",
			Cosa:        "tagliare l'erba",
			EntroQuando: time.Now().Add(3600 * time.Minute),
			Approvato:   true,
			Budget:      0,
			OreUomo:     4},
	}}

	tpl.ExecuteTemplate(w, "index.gohtml", data)
}

// StatoMinuta è true se tutti gli action points
// che compongono la minuta sono approvati
func (m Minuta) StatoMinuta() bool {
	for _, a := range m.AP {
		if a.Approvato == false {
			return false
		}
	}
	return true
}

// NewProcesso crea un nuovo processo.
func NewProcesso(titolo string) (p Processo, err error) {
	// Verifica che non esistano processi con lo stesso nome.
	processi, err := GetAllProcessi()
	if err != nil {
		return Processo{}, err
	}
	for _, processo := range processi {
		if processo.Titolo == titolo {
			return Processo{}, fmt.Errorf("titolo \"%s\" già esistente con id %s", titolo, processo.Id)
		}
	}
	p.Id = fmt.Sprintf("%x", md5.Sum([]byte(titolo)))
	p.Titolo = titolo
	p.Versione = 1
	p.Status = Nuovo
	p.Created_at = time.Now()
	p.Updated_at = time.Now()
	err = p.Save()
	return p, err
}

// GetAttivita recupera il processo con id.
func GetAttivita(id string) (Attivita, error) {
	var attivita Attivita
	err := a.Find(bson.M{"id": id}).One(&attivita)
	if err != nil {
		log.Printf("GetAttivita per id: %s in errore: %v \n", id, err)
	}
	return attivita, err
}

// GetProcesso recupera il processo con id.
func GetProcesso(id string) (Processo, error) {
	var p Processo
	err := c.Find(bson.M{"id": id}).One(&p)
	if err != nil {
		log.Printf("GetProcesso per id: %s in errore: %v \n", id, err)
	}
	return p, err
}

// GetAllProcessi recupera tutti i processi.
func GetAllProcessi() ([]Processo, error) {
	var processi []Processo
	err := c.Find(nil).All(&processi)
	if err != nil {
		log.Printf("GetAllProcessi in errore: %v \n", err)
	}
	return processi, err
}

// GetAllAttivita recupera tutti i processi.
func GetAllAttivita() ([]Attivita, error) {
	var attivita []Attivita
	err := a.Find(nil).All(&attivita)
	if err != nil {
		log.Printf("GetAllAttivita in errore: %v \n", err)
	}
	return attivita, err
}

// UpdateProcesso modifica un processo.
func UpdateProcesso(id string, p *Processo) {
	p.Id = id
	p.Versione++
	p.Updated_at = time.Now()

	c.Update(bson.M{"id": id}, &p)
}

// DeleteProcesso cancella un processo dal db.
func DeleteProcesso(id string) error {
	err := c.Remove(bson.M{"id": id})
	if err != nil {
		log.Print(err)
	}
	return err
}

// deleteAllProcessi cancella tutti i processi.
func deleteAllProcessi() error {
	var processi []Processo
	err := c.Find(nil).All(&processi)
	if err != nil {
		log.Printf("GetAllProcessi in errore: %v \n", err)
	}
	for _, p := range processi {
		c.Remove(bson.M{"id": p.Id})
	}
	return err
}

// deleteAllAttivita cancella tuttele attivita.
func deleteAllAttivita() error {
	var attivitas []Attivita
	err := a.Find(nil).All(&attivitas)
	if err != nil {
		log.Printf("deleteAllAttivita in errore: %v \n", err)
	}
	a.RemoveAll(nil)
	for _, attivita := range attivitas {
		a.Remove(bson.M{"id": attivita.Id})
	}
	return err
}
