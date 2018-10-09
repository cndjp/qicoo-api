package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/httprouter"
	"github.com/cndjp/qicoo-api/src/mysqlib"
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
	app.Version(fmt.Sprint("qicoo-api version: ", version))
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// TODO
	}

	// Redisの接続情報を与える
	pool.RedisPool = pool.NewRedisPool(os.Getenv("REDIS_URL"))

	// MySQLの接続情報を与える
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"

	mysqlib.SetConnectValue(user, password, protocol, dbname)
}

func main() {
	r := httprouter.MakeRouter(handler.QuestionCreateHandler, handler.QuestionListHandler, handler.QuestionDeleteHandler)

	logrus.Fatal(http.ListenAndServe(":8080", r))
}
