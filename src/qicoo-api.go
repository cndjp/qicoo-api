package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/httprouter"
	"github.com/cndjp/qicoo-api/src/pool"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
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

	pool.RedisPool = pool.NewRedisPool()
}

func main() {
	r := httprouter.MakeRouter(handler.QuestionCreateHandler, handler.QuestionListHandler)

	logrus.Fatal(http.ListenAndServe(":8080", r))
}
