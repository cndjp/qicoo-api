package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/handler"
	"github.com/cndjp/qicoo-api/src/httprouter"
)

func TestMain(t *testing.T) {
	const createQuestionMsg = "hello createQuestionFunc"
	const listQuestionMsg = "hello listQuestionFunc"
	const deleteQuestionMsg = "hello deleteQuestionFunc"
	const likeQuestionMsg = "hello likeQuestionFunc"
	const livenessMsg = "hello livenessFunc"
	const readinessMsg = "hello readinessFunc"

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
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, livenessMsg)
		},
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, readinessMsg)
		},
		nil)

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

	/* liveness */
	mockLiveReq := httptest.NewRequest("GET", "/liveness", nil)
	mockLiveRec := httptest.NewRecorder()

	r.ServeHTTP(mockLiveRec, mockLiveReq)

	if !reflect.DeepEqual(livenessMsg, mockLiveRec.Body.String()) {
		t.Errorf("expected %q to eq %q", livenessMsg, mockLiveRec.Body.String())
	}

	/* readiness */
	mockReadReq := httptest.NewRequest("GET", "/readiness", nil)
	mockReadRec := httptest.NewRecorder()

	r.ServeHTTP(mockReadRec, mockReadReq)

	if !reflect.DeepEqual(readinessMsg, mockReadRec.Body.String()) {
		t.Errorf("expected %q to eq %q", readinessMsg, mockReadRec.Body.String())
	}
}

func TestCORS(t *testing.T) {
	/* Preflight request (CORS) */

	tests := []struct {
		header        string
		expectedValue string
	}{
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Headers", "Content-Type"},
		{"Access-Control-Allow-Methods", "POST, PUT, DELETE"},
		{"Access-Control-Max-Age", "86400"},
	}

	r := httprouter.MakeRouter(
		nil, nil, nil, nil, nil, nil, handler.CORSPreflightHandler)

	mockPreCreateReq := httptest.NewRequest("OPTIONS", "/v1/mock/questions", nil)
	mockPreCreateRec := httptest.NewRecorder()
	r.ServeHTTP(mockPreCreateRec, mockPreCreateReq)

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if mockPreCreateRec.Header().Get(tt.header) != tt.expectedValue {
				t.Errorf("Prepost: expected %q to eq %q", tt.expectedValue, mockPreCreateRec.Header().Get(tt.header))
			}
		})
	}

	// Browsers send a preflight request to the same path before also PUT and DELETE request.
	mockPreDeleteReq := httptest.NewRequest("OPTIONS", "/v1/mock/questions/00000000-0000-4000-0000-000000000000", nil)
	mockPreDeleteRec := httptest.NewRecorder()
	r.ServeHTTP(mockPreDeleteRec, mockPreDeleteReq)
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if mockPreDeleteRec.Header().Get(tt.header) != tt.expectedValue {
				t.Errorf("Preput/delete: expected %q to eq %q", tt.expectedValue, mockPreDeleteRec.Header().Get(tt.header))
			}
		})
	}

}
