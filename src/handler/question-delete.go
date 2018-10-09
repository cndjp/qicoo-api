package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/cndjp/qicoo-api/src/mysqlib"
	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// QuestionDeleteHandler Questionの削除用 関数
func QuestionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれているパラメータを取得
	vars := mux.Vars(r)

	// 削除用のQuestion strictを生成 (GORPで使用するため)
	var q *Question
	q = new(Question)
	q.ID = vars["question_id"]
	q.EventID = vars["event_id"]

	/* DB connection 取得 */
	var m *gorp.DbMap
	db, err := mysqlib.InitMySQL()
	if err != nil {
		logrus.Error(err)
		return
	}

	m = mysqlib.MappingDBandTable(db)
	m.AddTableWithName(Question{}, "questions")
	defer m.Db.Close()

	if err != nil {
		logrus.Error(err)
		return
	}

	// DBからQuestionを削除
	err = QuestionDeleteDB(m, q)
	if err != nil {
		logrus.Error(err)
		return
	}

	// RedisからQurstionを削除
}

// QuestionDeleteDB DBからQuestionを削除する
func QuestionDeleteDB(m *gorp.DbMap, q *Question) error {
	// Tracelogの設定
	m.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	// delete実行
	_, err := m.Delete(q)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}
