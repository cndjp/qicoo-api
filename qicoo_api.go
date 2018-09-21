package main

import (
	_ "fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	http.ListenAndServe(":8080", r)
}
