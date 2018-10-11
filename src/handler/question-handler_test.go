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

var mockQuestion = handler.Question{
	ID:        "00000000-0000-0000-0000-000000000000",
	Object:    "question",
	Username:  "anonymous",
	EventID:   testEventID,
	ProgramID: "1",
	Comment:   "I am mock",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Like:      100000,
}

var mockQLMuxVars = handler.QuestionListMuxVars{
	EventID: testEventID,
	Start:   1,
	End:     100,
	Sort:    "created_at",
	Order:   "asc",
}

var mockQCMuxVars = handler.QuestionCreateMuxVars{
	EventID: testEventID,
}

var mockQDMuxVars = handler.QuestionDeleteMuxVars{
	EventID: testEventID,
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
	// DB and Table
	dbmap, _ := mockInitMySQL("") //DBが存在していないので、この時点ではDB名は指定しない
	createDBandTable(dbmap)

	// Rows
	dbmap, _ = mockInitMySQL("qicoo") //DB作成後はDB名を指定し直す必要がある
	generateMysqlTestdata(dbmap, mockQuestion)

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
func mockInitMySQL(dbnameS string) (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := "root"
	password := "my-secret-pw"
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
	dbmap, err := mockInitMySQL("qicoo")

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
