package handler

import (
	"net/http"
)

// LivenessHandler LivenessProbe用function
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
