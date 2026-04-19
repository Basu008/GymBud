package handler

import (
	"encoding/json"
	"net/http"
)

func BindJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
