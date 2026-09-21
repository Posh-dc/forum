package backEnd

import (
	"encoding/json"
	"net/http"
)

/* Resgistration response*/

type RegRes struct {
	Message   string
	ElementId string
}

func RegistrationResponse(w http.ResponseWriter, status int, message string, elementID string) {


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(RegRes{
		Message:   message,
		ElementId: elementID,
	})
}
