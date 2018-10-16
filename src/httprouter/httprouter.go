package httprouter

import (
	"net/http"

	"github.com/gorilla/mux"
)

// MakeRouter muxのroute設定用関数
func MakeRouter(questionCreateFunc, questionListFunc, questionDeleteFunc, questionLikeFunc, livenessFunc, readinessFunc func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter()

	// route QuestionCreate
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("POST").
		HandlerFunc(questionCreateFunc)

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

	// route QuestionLike
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions/{question_id:[a-zA-Z0-9-_]+}/like").
		Methods("PUT").
		HandlerFunc(questionLikeFunc)

	// route LivenessProbe
	r.Path("/liveness").
		Methods("GET").
		HandlerFunc(livenessFunc)

	// route ReadinessProbe
	r.Path("/readiness").
		Methods("GET").
		HandlerFunc(readinessFunc)

	return r
}
