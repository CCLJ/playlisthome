package handlers

import (
	"encoding/json"
	"net/http"
)

// NotImplemented is a placeholder for routes that haven't been built yet.
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "not implemented yet"})
}
