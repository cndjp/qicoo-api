package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/httprouter"
)

func TestMain(t *testing.T) {
	const createQuestionMsg = "hello createQuestionFunc"
	const listQuestionMsg = "hello listQuestionFunc"
	const deleteQuestionMsg = "hello deleteQuestionFunc"
	const likeQuestionMsg = "hello likeQuestionFunc"

	r := httprouter.MakeRouter(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, createQuestionMsg)
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, listQuestionMsg)
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, deleteQuestionMsg)
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, likeQuestionMsg)
		})

	/* CreateQuestion */
	mockCreateReq := httptest.NewRequest("POST", "/v1/mock/questions", nil)
	mockCreateRec := httptest.NewRecorder()

	r.ServeHTTP(mockCreateRec, mockCreateReq)

	if !reflect.DeepEqual(createQuestionMsg, mockCreateRec.Body.String()) {
		t.Errorf("expected %q to eq %q", createQuestionMsg, mockCreateRec.Body.String())
	}

	/* ListQuestion */
	mockListReq := httptest.NewRequest("GET", "/v1/mock/questions?start=1&end=100&sort=created_at&order=asc", nil)
	mockListRec := httptest.NewRecorder()

	r.ServeHTTP(mockListRec, mockListReq)

	if !reflect.DeepEqual(listQuestionMsg, mockListRec.Body.String()) {
		t.Errorf("expected %q to eq %q", listQuestionMsg, mockListRec.Body.String())
	}

	/* DeleteQuestion */
	mockDeleteReq := httptest.NewRequest("DELETE", "/v1/mock/questions/questionDummyId", nil)
	mockDeleteRec := httptest.NewRecorder()

	r.ServeHTTP(mockDeleteRec, mockDeleteReq)

	if !reflect.DeepEqual(deleteQuestionMsg, mockDeleteRec.Body.String()) {
		t.Errorf("expected %q to eq %q", deleteQuestionMsg, mockDeleteRec.Body.String())
	}

	/* LikeQuestion */
	mockLikeReq := httptest.NewRequest("PUT", "/v1/mock/questions/questionDummyId/like", nil)
	mockLikeRec := httptest.NewRecorder()

	r.ServeHTTP(mockLikeRec, mockLikeReq)

	if !reflect.DeepEqual(likeQuestionMsg, mockLikeRec.Body.String()) {
		t.Errorf("expected %q to eq %q", likeQuestionMsg, mockLikeRec.Body.String())
	}
}
