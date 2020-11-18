package session

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// HTTPrequest used to setup a session
type HTTPrequest struct {
	Device string `json:"device"`
}

// HTTPresponse used to setup a session
type HTTPresponse struct {
	SessionID uuid.UUID `json:"sessionID,omitempty"`
	TunnelID  uuid.UUID `json:"tunnelID,omitempty"`
	ErrorMsg  string    `json:"error,omitempty"`
}

// RespondSuccess indicates successful session setup
func RespondSuccess(w http.ResponseWriter, res HTTPresponse) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RespondError indicates unsuccessful session setup
func RespondError(w http.ResponseWriter, e error, status int) {
	res := HTTPresponse{
		ErrorMsg: e.Error(),
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
