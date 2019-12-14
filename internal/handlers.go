package internal

import (
	"encoding/json"
	"net/http"
)

func (s *server) Handler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(request)
	defer r.Body.Close()
	response := new(Response)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.Error = ErrInvalidRequestBody
		responseData, err := json.Marshal(response)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write(responseData); err != nil {
			panic(err)
		}
		return
	}
}