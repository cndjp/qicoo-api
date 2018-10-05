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
	const createMsg = "hello createFunc"
	const listMsg = "hello listFunc"

	r := httprouter.MakeRouter(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, createMsg)
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, listMsg)
		})

	mockCreateReq := httptest.NewRequest("POST", "/v1/mock/questions", nil)
	mockCreateRec := httptest.NewRecorder()

	r.ServeHTTP(mockCreateRec, mockCreateReq)

	if !reflect.DeepEqual(createMsg, mockCreateRec.Body.String()) {
		t.Errorf("expected %q to eq %q", createMsg, mockCreateRec.Body.String())
	}

	mockListReq := httptest.NewRequest("GET", "/v1/mock/questions?start=1&end=100&sort=created_at&order=asc", nil)
	mockListRec := httptest.NewRecorder()

	r.ServeHTTP(mockListRec, mockListReq)

	if !reflect.DeepEqual(listMsg, mockListRec.Body.String()) {
		t.Errorf("expected %q to eq %q", listMsg, mockListRec.Body.String())
	}
}
