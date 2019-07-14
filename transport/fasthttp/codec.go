package fasthttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unsafe"

	"github.com/valyala/fasthttp"
)

var (
	// IngoreContentType Content-Type header will be ignored when decoding the body
	IngoreContentType bool
)

// NewObjectFunc acquire a clean object
type NewObjectFunc func() (object interface{})

// ReleaseObjectFunc reset and release a object
type ReleaseObjectFunc func(object interface{})

// EncodeRequestFunc encodes the passed request object into the HTTP request
// object. It's designed to be used in HTTP clients, for client-side
// endpoints. One straightforward EncodeRequestFunc could be something that JSON
// encodes the object directly to the request body.
type EncodeRequestFunc func(ctx context.Context, req *fasthttp.Request, request interface{}) (err error)

// DecodeRequestFunc extracts a user-domain request object from an HTTP
// request object. It's designed to be used in HTTP servers, for server-side
// endpoints. One straightforward DecodeRequestFunc could be something that
// JSON decodes from the request body to the concrete request type.
type DecodeRequestFunc func(ctx context.Context, req *fasthttp.Request, request interface{}) (err error)

// EncodeResponseFunc encodes the passed response object to the HTTP response
// writer. It's designed to be used in HTTP servers, for server-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly to the response body.
type EncodeResponseFunc func(ctx context.Context, resp *fasthttp.Response, resonse interface{}) (err error)

// DecodeResponseFunc extracts a user-domain response object from an HTTP
// response object. It's designed to be used in HTTP clients, for client-side
// endpoints. One straightforward DecodeResponseFunc could be something that
// JSON decodes from the response body to the concrete response type.
type DecodeResponseFunc func(ctx context.Context, resp *fasthttp.Response, response interface{}) (err error)

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Request body. Many JSON-over-HTTP services can use it as
// a sensible default. If the request implements Headerer, the provided headers
// will be applied to the request.
func EncodeJSONRequest(_ context.Context, r *fasthttp.Request, request interface{}) error {
	r.Header.SetContentType("application/json; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	return json.NewEncoder(r.BodyWriter()).Encode(request)
}

// DecodeJSONRequest is an DecodeRequestFunc that deserializes the response as a
// JSON object from the Request body. Many JSON-over-HTTP services can use it as
// a sensible default.
func DecodeJSONRequest(_ context.Context, r *fasthttp.Request, request interface{}) error {
	if !IngoreContentType {
		contentType := strings.Split(b2s(r.Header.ContentType()), ";")[0]

		if contentType != "application/json" {
			return errors.New("Content-Type not's application/json")
		}
	}

	return json.Unmarshal(r.Body(), request)
}

// EncodeJSONResponse is a EncodeResponseFunc that serializes the response as a
// JSON object to the ResponseWriter. Many JSON-over-HTTP services can use it as
// a sensible default. If the response implements Headerer, the provided headers
// will be applied to the response. If the response implements StatusCoder, the
// provided StatusCode will be used instead of 200.
func EncodeJSONResponse(_ context.Context, resp *fasthttp.Response, response interface{}) error {
	resp.Header.SetContentType("application/json; charset=utf-8")
	if headerer, ok := response.(Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				resp.Header.Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	resp.SetStatusCode(code)
	if code == http.StatusNoContent {
		return nil
	}
	return json.NewEncoder(resp.BodyWriter()).Encode(response)
}

// DecodeJSONResponse is an DecodeResponseFunc that deserializes the response as a
// JSON object from the Response body. Many JSON-over-HTTP services can use it as
// a sensible default.
func DecodeJSONResponse(_ context.Context, resp *fasthttp.Response, response interface{}) error {
	if !IngoreContentType {
		contentType := strings.Split(b2s(resp.Header.ContentType()), ";")[0]

		if contentType != "application/json" {
			return errors.New("Content-Type not's application/json")
		}
	}

	// The fasthttp client always reads the whole body into memory before returning to the program.
	// https://github.com/valyala/fasthttp/issues/246
	return json.Unmarshal(resp.Body(), response)
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
