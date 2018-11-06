package mysqlib

import (
	"database/sql"
	"os"
	"strings"

	"github.com/cndjp/qicoo-api/src/loglib"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

// GetMySQLdbmap Dbmapの取得
func GetMySQLdbmap() (dbmap *gorp.DbMap, err error) {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}

	return dbmap, nil
}

// InitDB DBの初期設定。DatabaseやTableが存在しない場合は作成する
func InitDB() error {
	sugar := loglib.GetSugar()
	defer sugar.Sync()

	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := ""
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		sugar.Error(err)
		return err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}

	// DATABASEの作成 (DATABASEが存在するか確認する良い方法がなかったため、CREATEを投げている)
	_, err = dbmap.Exec("CREATE DATABASE qicoo;")

	if err != nil {
		errmsg := err.Error()

		// DABABASEが存在しているerrmsgの場合は、正常状態とする
		if strings.Contains(errmsg, "Can't create database 'qicoo'; database exists") {
			sugar.Info("qicoo DATABASE exists")
		} else {
			sugar.Error(err)
			return err
		}
	}

	/* Tableの作成 */
	_, err = dbmap.Exec("CREATE TABLE qicoo.questions (" +
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

	if err != nil {
		errmsg := err.Error()

		// TABLEが存在しているerrmsgの場合は、正常状態とする
		if strings.Contains(errmsg, "Table 'questions' already exists") {
			sugar.Info("questions TABLE exists")
		} else {
			sugar.Error(err)
			return err
		}
	}

	return nil
}
