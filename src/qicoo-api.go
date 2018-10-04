package main

import (
	"fmt"
	"os"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/httprouter"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version string

var (
	app = kingpin.New("qicoo-api", "This application is Qicoo's Backend API")

	verbose = app.Flag("verbose", "Run verbose mode").Default("false").Short('v').Bool()
)

func init() {
	app.HelpFlag.Short('h')
	app.Version(fmt.Sprint("dntk's version: ", version))
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// TODO
	}
}

func main() {
	//r := mux.NewRouter()

	// RedisPoolの初期化初期設定
	//var p = handler.NewRedisPool()

	// route QuestionCreate
	//r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
	//Methods("POST").
	//HandlerFunc(handler.QuestionCreateHandler)

	// route QuestionList
	//r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
	// 	Methods("GET").
	// 	Queries("start", "{start:[0-9]+}").
	// 	Queries("end", "{end:[0-9]+}").
	// 	Queries("sort", "{sort:[a-zA-Z0-9-_]+}").
	// 	Queries("order", "{order:[a-zA-Z0-9-_]+}").
	// 	HandlerFunc(handler.QuestionListHandler)

	// http.ListenAndServe(":8080", r)

	httprouter.Run(handler.QuestionCreateHandler, handler.QuestionListHandler)
}
