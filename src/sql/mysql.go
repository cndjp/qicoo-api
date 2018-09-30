package sql

import (
	goSQL "database/sql"
	"os"

	"github.com/go-gorp/gorp"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DBMap struct {
	Map *gorp.DbMap
}

func envLoad() {
	err := godotenv.Load("./.env")
	if err != nil {
		logrus.Fatal("Error loading .env file")

	}
}

// InitMySQLDB 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
func (m *DBMap) InitMySQLDB() (err error) {
	envLoad()

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
