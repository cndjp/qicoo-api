package mysqlib

import (
	"database/sql"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

var user string
var password string
var protocol string
var dbname string

// InitMySQL DBの初期設定
func InitMySQL() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	//user := os.Getenv("DB_USER")
	//password := os.Getenv("DB_PASSWORD")
	//protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	//dbname := "qicoo"
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

// SetConnectValue mysqlibパッケージのグローバル変数に定義。テストコードのために使用する
func SetConnectValue(userS, passwordS, protocolS, dbnameS string) {
	user = userS
	password = passwordS
	protocol = protocolS
	dbname = dbnameS
}

// MappingDBandTable 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
//func MappingDBandTable(db *sql.DB) *gorp.DbMap {
//	// structの構造体とDBのTableを紐づける
//	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
//
//	return dbmap
//}
