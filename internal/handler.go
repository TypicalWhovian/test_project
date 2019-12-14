package internal

import (
	"encoding/json"
	"net/http"
)

type Request struct {
	requestId   string `json:"requestId"`
	requestType string `json:"type"`
}

type Err struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type Response struct {
	Data  interface{} `json:"data"`
	Error *Err        `json:"error"`
}

var (
	ErrInvalidRequestBody = &Err{"request body is invalid", 1}
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
