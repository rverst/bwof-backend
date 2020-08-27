package helper

import (
	"encoding/json"
	"github.com/rverst/bwof-backend/pkg/models"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, model interface{}) (int, error) {
	d, err := json.Marshal(model)
	if err != nil {
		return 0, err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return w.Write(d)
}

func WriteError(w http.ResponseWriter, status int, message string) (int, error) {
	res := models.Response{
		Status: status,
		Error:  message,
	}
	buf, err := json.Marshal(res)
	if err != nil {
		return 0, err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return w.Write(buf)
}
