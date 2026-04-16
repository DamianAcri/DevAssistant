package handlers

import (
	"encoding/json"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json") //json res
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}