package handler

import "net/http"

// CorsPreflightHandler handles preflight OPTIONS request before cross-origin API access.
func CORSPreflightHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, DELETE")
	w.Header().Set("Access-Control-Max-Age", "86400")
}
