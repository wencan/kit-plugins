package fasthttp

import (
	"context"

	"github.com/valyala/fasthttp"
)

// NewObjectFunc new a clear object
type NewObjectFunc func() (object interface{})

// ReleaseObjectFunc clear and release a object
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
