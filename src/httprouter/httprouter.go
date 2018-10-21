package httprouter

import (
	"net/http"

	"github.com/gorilla/mux"
)

// MakeRouter muxのroute設定用関数
func MakeRouter(questionCreateFunc, questionListFunc, questionDeleteFunc, corsPrelightFunc func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("POST").
		HandlerFunc(questionCreateFunc)
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("OPTIONS").
		HandlerFunc(corsPrelightFunc)

	// route QuestionList
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("GET").
		Queries("start", "{start:[0-9]+}").
		Queries("end", "{end:[0-9]+}").
		Queries("sort", "{sort:[a-zA-Z0-9-_]+}").
		Queries("order", "{order:[a-zA-Z0-9-_]+}").
		HandlerFunc(questionListFunc)

	// route QuestionDelete
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions/{question_id:[a-zA-Z0-9-_]+}").
		Methods("DELETE").
		HandlerFunc(questionDeleteFunc)
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions/{question_id:[a-zA-Z0-9-_]+}").
		Methods("OPTIONS").
		HandlerFunc(corsPrelightFunc)

	return r
}
