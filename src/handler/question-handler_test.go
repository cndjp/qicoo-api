// questionHandlerに関するテストコードの共通部分定義

package handler_test

import (
	"database/sql"
	"fmt"
	_ "log"
	"os"
	"testing"
	"time"

	"github.com/cndjp/qicoo-api/src/handler"
	_ "github.com/cndjp/qicoo-api/src/mysqlib"
	"github.com/go-gorp/gorp"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

var mockRedisPool *redis.Pool

const testEventID = "testEventID"

var mockQLMuxVars = handler.QuestionListMuxVars{
	EventID: testEventID,
	Start:   1,
	End:     100,
	Sort:    "created_at",
	Order:   "desc",
}

var mockQCMuxVars = handler.QuestionCreateMuxVars{
	EventID: testEventID,
}

var mockQDMuxVars = handler.QuestionDeleteMuxVars{
	EventID: testEventID,
}

var mockQLikeMuxVars = handler.QuestionLikeMuxVars{
	EventID:    testEventID,
	QuestionID: "00000000-0000-0000-0000-000000000000",
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
	/* Redisの接続情報設定 */
	mockRedisPool = mockNewRedisPool()

	/* MySQLのテストデータ格納 */
	var dbtestpasswd string
	if isTravisEnv() {
		dbtestpasswd = ""
	} else {
		dbtestpasswd = "my-secret-pw"
	}

	// DB and Table
	dbmap, _ := mockInitMySQL(dbtestpasswd, "") //DBが存在していないので、この時点ではDB名は指定しない
	createDBandTable(dbmap)

	// Rows
	dbmap, _ = mockInitMySQL(dbtestpasswd, "qicoo") //DB作成後はDB名を指定し直す必要がある
	generateMysqlTestdata(dbmap, getMockQuestion())

	return m.Run()
}

func flushallRedis(conn redis.Conn) {
	if _, err := conn.Do("FLUSHALL"); err != nil {
		fmt.Println(err)
	}
}

// mockNewRedisPool Mock用のRedisPoolを生成
func mockNewRedisPool() *redis.Pool {
	// idle connection limit:3    active connection limit:1000
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", "127.0.0.1:6379") },
	}
}

// mockRedisManager RedisConnectionInterfaceの実装
type mockRedisManager struct {
}

// GetRedisConnection RedisConnectionの取得
func (rm *mockRedisManager) GetRedisConnection() (conn redis.Conn) {
	return mockRedisPool.Get()
}

// mockInitMySQL Mock用DBの初期設定(DockerContainer)
func mockInitMySQL(dbtestpasswd string, dbnameS string) (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := "root"
	password := dbtestpasswd
	protocol := "tcp(127.0.0.1)"
	dbname := dbnameS
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	dbmap.AddTableWithName(handler.Question{}, "questions")

	return dbmap, nil
}

//MySQLManager MySQLDbmapInterfaceの実装
type mockMySQLManager struct {
}

// GetMySQLdbmap DBのdbmapを取得
func (mm *mockMySQLManager) GetMySQLdbmap() *gorp.DbMap {
	var dbtestpasswd string
	if isTravisEnv() {
		dbtestpasswd = ""
	} else {
		dbtestpasswd = "my-secret-pw"
	}

	dbmap, err := mockInitMySQL(dbtestpasswd, "qicoo")

	if err != nil {
		logrus.Error(err)
		return nil
	}

	dbmap.AddTableWithName(handler.Question{}, "questions")
	return dbmap
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
	err := dbmap.Insert(&question)

	if err != nil {
		logrus.Error(err)
	}
}

// getMockQuestion
func getMockQuestion() handler.Question {
	var loc, _ = time.LoadLocation("Asia/Tokyo")
	mocktime, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-10-01 12:12:12", loc)

	var mq handler.Question
	mq.ID = "00000000-0000-0000-0000-000000000000"
	mq.Object = "question"
	mq.Username = "anonymous"
	mq.EventID = testEventID
	mq.ProgramID = "1" // 未実装のため1固定で生成
	mq.Comment = "I am Mock"
	mq.Like = 0
	mq.UpdatedAt = mocktime
	mq.CreatedAt = mocktime

	return mq
}
