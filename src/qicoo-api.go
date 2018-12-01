package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/httprouter"
	"github.com/cndjp/qicoo-api/src/loglib"
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
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	app.HelpFlag.Short('h')
	app.Version(fmt.Sprint("qicoo-api version: ", version))
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// TODO
	}

	// Redis Poolを生成
	pool.RedisPool = pool.NewRedisPool()

	// DB, Tableが存在しない場合は作成する
	err := mysqlib.InitDB()

	if err != nil {
		sugar.Error(err)
		return
	}
}

func main() {
	defer mysqlib.CloseDB()

	r := httprouter.MakeRouter(
		handler.QuestionCreateHandler,
		handler.QuestionListHandler,
		handler.QuestionDeleteHandler,
		handler.QuestionLikeHandler,
		handler.LivenessHandler,
		handler.ReadinessHandler,
		handler.CORSPreflightHandler)

	logrus.Fatal(http.ListenAndServe(":8080", r))
}
