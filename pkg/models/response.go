package models

import "github.com/google/uuid"

type Response struct {
	Status int         `json:"status"`
	Error  string      `json:"error"`
	Data   interface{} `json:"data"`
}

type UploadResponse struct {
	Id uuid.UUID `json:"id"`
}
