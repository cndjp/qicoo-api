package main

import (
	_ "bytes"
	"encoding/json"
	_ "fmt"
	_ "log"
	"net/http"

	"github.com/gorilla/mux"
)

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Username  string `json:"username"`
	EventID   string `json:"event_id"`
	ProgramID string `json:"program_id"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
	UpdateAt  string `json:"update_at"`
	Like      int    `json:"like"`
}

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var question Question
	decoder.Decode(&question)
	w.Write([]byte("comment " + question.Comment))
	//question := &Question{ID: "testid"}

	//var buf bytes.Buffer
	//enc := json.NewEncoder(&buf)
	//if err := enc.Encode(question); err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println(buf.String())
}

func main() {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	http.ListenAndServe(":8080", r)
}
