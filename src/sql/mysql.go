package sql

import (
	goSQL "database/sql"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/pkg/errors"
)

type DBMap struct {
	Map *gorp.DbMap
}

// InitMySQLDB 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
func (m *DBMap) InitMySQLDB() (err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := goSQL.Open(dbms, connect)

	if err != nil {
		return errors.Wrap(err, "error on MySQL initialization.")
	}

	// structの構造体とDBのTableを紐づける
	m.Map = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	// dbmap.AddTableWithName(handler.Question{}, "questions")

	return nil
}
