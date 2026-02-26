package lnbot

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error response from the LnBot API.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("lnbot: %s (status %d)", e.Message, e.StatusCode)
}

// BadRequestError is returned for 400 responses.
type BadRequestError struct{ *APIError }

// UnauthorizedError is returned for 401 responses.
type UnauthorizedError struct{ *APIError }

// ForbiddenError is returned for 403 responses.
type ForbiddenError struct{ *APIError }

// NotFoundError is returned for 404 responses.
type NotFoundError struct{ *APIError }

// ConflictError is returned for 409 responses.
type ConflictError struct{ *APIError }

func (e *BadRequestError) Unwrap() error   { return e.APIError }
func (e *UnauthorizedError) Unwrap() error { return e.APIError }
func (e *ForbiddenError) Unwrap() error    { return e.APIError }
func (e *NotFoundError) Unwrap() error     { return e.APIError }
func (e *ConflictError) Unwrap() error     { return e.APIError }

func parseAPIError(statusCode int, body []byte) error {
	msg := parseErrorMessage(body)
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	base := &APIError{StatusCode: statusCode, Message: msg, Body: string(body)}
	switch statusCode {
	case http.StatusBadRequest:
		return &BadRequestError{base}
	case http.StatusUnauthorized:
		return &UnauthorizedError{base}
	case http.StatusForbidden:
		return &ForbiddenError{base}
	case http.StatusNotFound:
		return &NotFoundError{base}
	case http.StatusConflict:
		return &ConflictError{base}
	default:
		return base
	}
}

func parseErrorMessage(body []byte) string {
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) != nil {
		return ""
	}
	if parsed.Message != "" {
		return parsed.Message
	}
	return parsed.Error
}
