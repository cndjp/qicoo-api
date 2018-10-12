package mysqlib

import (
	"database/sql"
	"os"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// InitMySQL DBの初期設定
func InitMySQL() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}

	return dbmap, nil
}
