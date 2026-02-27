package lnbot

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	e := &APIError{StatusCode: 500, Message: "fail", Body: "{}"}
	want := "lnbot: fail (status 500)"
	if got := e.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAPIError_Fields(t *testing.T) {
	e := &APIError{StatusCode: 400, Message: "bad", Body: `{"error":"bad"}`}
	if e.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", e.StatusCode)
	}
	if e.Body != `{"error":"bad"}` {
		t.Errorf("Body = %q", e.Body)
	}
}

func TestTypedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"BadRequest", &BadRequestError{&APIError{StatusCode: 400, Message: "bad"}}, 400},
		{"Unauthorized", &UnauthorizedError{&APIError{StatusCode: 401, Message: "unauth"}}, 401},
		{"Forbidden", &ForbiddenError{&APIError{StatusCode: 403, Message: "forbidden"}}, 403},
		{"NotFound", &NotFoundError{&APIError{StatusCode: 404, Message: "not found"}}, 404},
		{"Conflict", &ConflictError{&APIError{StatusCode: 409, Message: "conflict"}}, 409},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var apiErr *APIError
			if !errors.As(tt.err, &apiErr) {
				t.Fatal("expected to unwrap to *APIError")
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
		})
	}
}

func TestParseErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"message field", `{"message":"invalid amount"}`, "invalid amount"},
		{"error field", `{"error":"bad input"}`, "bad input"},
		{"message takes precedence", `{"message":"msg","error":"err"}`, "msg"},
		{"invalid json", `not json`, ""},
		{"no fields", `{"detail":"something"}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseErrorMessage([]byte(tt.body))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		check  func(error) bool
	}{
		{"BadRequest", 400, func(e error) bool { var t *BadRequestError; return errors.As(e, &t) }},
		{"Unauthorized", 401, func(e error) bool { var t *UnauthorizedError; return errors.As(e, &t) }},
		{"Forbidden", 403, func(e error) bool { var t *ForbiddenError; return errors.As(e, &t) }},
		{"NotFound", 404, func(e error) bool { var t *NotFoundError; return errors.As(e, &t) }},
		{"Conflict", 409, func(e error) bool { var t *ConflictError; return errors.As(e, &t) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.status, []byte(`{"message":"test"}`))
			if !tt.check(err) {
				t.Errorf("status %d: got %T, want matching typed error", tt.status, err)
			}
		})
	}

	t.Run("unknown status returns *APIError", func(t *testing.T) {
		err := parseAPIError(500, []byte(`{"message":"fail"}`))
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatal("expected *APIError")
		}
		if apiErr.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
		}
	})
}
