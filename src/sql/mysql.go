package sql

import (
	"database/sql"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/pkg/errors"
)

// initMySQLDB 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
func InitMySQLDB() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		return nil, errors.Wrap(err, "error on initDb()")
	}

	// structの構造体とDBのTableを紐づける
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	// dbmap.AddTableWithName(handler.Question{}, "questions")

	return dbmap, nil
}
