package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/axamon/uddi"
	"github.com/gorilla/mux"
)

func main() {
	addr := ":8080"
	// r := http.NewServeMux()
	r := mux.NewRouter()

	r.HandleFunc("/", index)
	r.HandleFunc("/minuta/nuova", nuova)
	r.HandleFunc("/minuta/{ID}", getMinuta)

	http.ListenAndServe(addr, r)
}

func getMinuta(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["ID"])

}

func nuova(w http.ResponseWriter, r *http.Request) {

	id := uddi.Create()

	fmt.Fprintf(w, "Category: %v\n", id)


	t, err := template.ParseFiles("templates/nuova.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.ExecuteTemplate(w, "nuova.gohtml", nil)
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

	t, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Println(err)
	}
	t.ExecuteTemplate(w, "index.gohtml", data)
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
