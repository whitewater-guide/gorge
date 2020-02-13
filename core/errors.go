package core

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

type ErrorType int

type Error struct {
	Err error                  `json:"-"`
	Msg string                 `json:"error"`
	Ctx map[string]interface{} `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) With(key string, val interface{}) *Error {
	if e.Ctx == nil {
		e.Ctx = map[string]interface{}{}
	}
	e.Ctx[key] = val
	return e
}

func (e *Error) WithMap(m map[string]interface{}) *Error {
	if e.Ctx == nil {
		e.Ctx = map[string]interface{}{}
	}
	for k, v := range m {
		e.Ctx[k] = v
	}
	return e
}

func NewErr(err error, ctx ...interface{}) *Error {
	res := &Error{
		Msg: err.Error(),
	}
	if len(ctx) == 1 {
		ctxmap, ok := ctx[0].(map[string]interface{})
		if ok {
			res.Ctx = ctxmap
		}
	} else {
		res.Ctx = map[string]interface{}{}
	}
	return res
}

func WrapErr(err error, msg string, ctx ...interface{}) *Error {
	res := &Error{
		Err: err,
		Msg: msg,
	}
	if len(ctx) == 1 {
		ctxmap, ok := ctx[0].(map[string]interface{})
		if ok {
			res.Ctx = ctxmap
		}
	} else {
		res.Ctx = map[string]interface{}{}
	}

	// lift context up so it can be logged
	if e, ok := err.(*Error); ok {
		for k, v := range e.Ctx {
			res.Ctx[k] = v
		}
	}
	return res
}

type ErrorResponse struct {
	*Error
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status,omitempty"` // user-level status message
	ReqID          string `json:"request_id,omitempty"`
}

func NewErrorResponse(e error, message string, code int) *ErrorResponse {
	var resp *ErrorResponse
	if err, ok := e.(*Error); ok {
		resp = &ErrorResponse{Error: err}
	} else {
		resp = &ErrorResponse{
			Error: &Error{
				Msg: "internal server error",
				Err: e,
			},
		}
	}
	resp.HTTPStatusCode = code
	resp.StatusText = message
	return resp
}

func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}
