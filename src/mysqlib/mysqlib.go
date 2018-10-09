package mysqlib

import (
	"database/sql"
	"os"

	"github.com/go-gorp/gorp"
)

func InitMySQL() (db *sql.DB, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err = sql.Open(dbms, connect)

	return
}

// InitMySQLDB 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
func MappingDBandTable(db *sql.DB) *gorp.DbMap {
	// structの構造体とDBのTableを紐づける
	return &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
}
