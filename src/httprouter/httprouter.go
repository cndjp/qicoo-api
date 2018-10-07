package httprouter

import (
	"net/http"

	"github.com/gorilla/mux"
)

func MakeRouter(createFunc, listFunc func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("POST").
		HandlerFunc(createFunc)

	// route QuestionList
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("GET").
		Queries("start", "{start:[0-9]+}").
		Queries("end", "{end:[0-9]+}").
		Queries("sort", "{sort:[a-zA-Z0-9-_]+}").
		Queries("order", "{order:[a-zA-Z0-9-_]+}").
		HandlerFunc(listFunc)

	return r
}
