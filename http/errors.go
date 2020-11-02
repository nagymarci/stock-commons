package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

var UnknownError = "{\"message\":\"Uknown error\"}"

type HttpError interface {
	Status() int
}

type InternalServerError struct {
	err    string
	status int
}

func NewInternalServerError(msg string) error {
	return &InternalServerError{
		err:    msg,
		status: http.StatusInternalServerError,
	}
}

func (e *InternalServerError) Error() string {
	return e.err
}

func (e *InternalServerError) Status() int {
	return e.status
}

type BadRequestError struct {
	err    string
	status int
}

func NewBadRequestError(msg string) error {
	return &BadRequestError{
		err:    msg,
		status: http.StatusBadRequest,
	}
}

func (e *BadRequestError) Error() string {
	return e.err
}

func (e *BadRequestError) Status() int {
	return e.status
}

type notFoundError struct {
	err    string
	status int
}

func NewNotFoundError(msg string) error {
	return &notFoundError{
		err:    msg,
		status: http.StatusNotFound,
	}
}

func (e *notFoundError) Error() string {
	return e.err
}

func (e *notFoundError) Status() int {
	return e.status
}

func HandleError(err error, w http.ResponseWriter) {
	statusCode := http.StatusInternalServerError
	if err, ok := interface{}(&err).(HttpError); ok {
		statusCode = err.Status()
	}
	message := "Failed to process request: " + err.Error()
	HandleErrorResponse(message, w, statusCode)
}

func HandleErrorResponse(msg string, w http.ResponseWriter, status int) {
	response := ErrorResponse{Message: msg}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, UnknownError, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}
