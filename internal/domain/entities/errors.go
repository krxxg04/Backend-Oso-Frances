package domain

import "net/http"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type ErrorResponse struct {
	Error  APIError   `json:"error"`
	Errors []APIError `json:"errors,omitempty"`
}

func NewError(code, message, field string) APIError {
	return APIError{Code: code, Message: message, Field: field}
}

func HTTPStatusForCode(code string) int {
	switch code {
	case "validation_error":
		return http.StatusBadRequest
	case "unauthorized":
		return http.StatusUnauthorized
	case "not_found":
		return http.StatusNotFound
	case "rate_limited":
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
