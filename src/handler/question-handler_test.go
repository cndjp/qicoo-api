// questionに関するテストコードの共通部分定義

package handler_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/mysqlib"
	"github.com/cndjp/qicoo-api/src/pool"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

var travisTestRedisConn redis.Conn
var internalTestRedisConn *redigomock.Conn

const testEventID = "testEventID"

var mockQuestion = handler.Question{
	ID:        "1",
	Object:    "question",
	Username:  "anonymous",
	EventID:   testEventID,
	ProgramID: "1",
	Comment:   "I am mock",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Like:      100000,
}

var mockMuxVars = handler.MuxVars{
	EventID: testEventID,
	Start:   1,
	End:     100,
	Sort:    "created_at",
	Order:   "asc",
}

func isTravisEnv() bool {
	if os.Getenv("TRAVIS") == "true" {
		return true
	}
	return false
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	if isTravisEnv() {
		conn, err := redis.Dial("tcp", "localhost:6379")
		if err != nil {
			log.Fatal(err)
		}

		travisTestRedisConn = conn
	} else {
		/* Redisの接続情報設定 */
		setLocalRedisPool(newLocalRedisPool())

		/* MySQLのテストデータ格納 */
		// DB and Table
		mysqlib.SetConnectValue("root", "my-secret-pw", "tcp(127.0.0.1)", "") //DBが存在していないので、この時点ではDB名は指定しない
		dbmap := handler.InitMySQLQuestion()
		createDBandTable(dbmap)

		// Rows
		mysqlib.SetConnectValue("root", "my-secret-pw", "tcp(127.0.0.1)", "qicoo") //DB作成後はDB名を指定し直す必要がある
		dbmap = handler.InitMySQLQuestion()
		generateMysqlTestdata(dbmap, mockQuestion)
	}

	return m.Run()
}

func flushallRedis(conn redis.Conn) {
	if _, err := conn.Do("FLUSHALL"); err != nil {
		fmt.Println(err)
	}
}

func newMockPool() *handler.RedisClient {
	m := new(handler.RedisClient)
	if isTravisEnv() {
		m.RedisConn = travisTestRedisConn
	} else {
		m.RedisConn = internalTestRedisConn
	}

	m.Vars = mockMuxVars

	return m
}

// newLocalRedisPool LocalのRedisで使用するRedisPoolを生成
func newLocalRedisPool() *redis.Pool {
	return pool.NewRedisPool("127.0.0.1:6379")
}

// handlerのRedisPoolにLocal用のRedisPoolを設定する
func setLocalRedisPool(redisPool *redis.Pool) {
	pool.RedisPool = redisPool
}

// createDBandTable MySQLにDatabaseとTableを作成
func createDBandTable(dbmap *gorp.DbMap) {
	/* DBの作成 */
	dbmap.Exec("CREATE DATABASE qicoo;")

	/* Tableの作成 */
	dbmap.Exec("CREATE TABLE qicoo.questions (" +
		"id varchar(36) PRIMARY KEY," +
		"object text," +
		"event_id text," +
		"program_id text," +
		"username text," +
		"comment text," +
		"like_count int(10)," +
		"created_at DATETIME," +
		"updated_at DATETIME," +
		"INDEX (event_id(40), program_id(40))" +
		");")
}

// generateMysqlTestdata MySQLのテストデータを生成して格納
func generateMysqlTestdata(dbmap *gorp.DbMap, question handler.Question) {
	/* データの挿入 */
	dbmap.Insert(&question)
}
