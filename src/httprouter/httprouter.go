package httprouter

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Run(createFunc, ListFunc func(w http.ResponseWriter, r *http.Request)) {
	r := mux.NewRouter()

	// RedisPoolの初期化初期設定
	//var p = handler.NewRedisPool()

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
		HandlerFunc(ListFunc)

	http.ListenAndServe(":8080", r)
}
