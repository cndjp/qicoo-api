package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID string `json:"id"`
}

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	question := &Question{ID: "testid"}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(question); err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.String())
}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	http.ListenAndServe(":8080", r)
}
