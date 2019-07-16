package fasthttp

import (
	"context"

	"github.com/valyala/fasthttp"
)

type contextKey int

const (
	// ContextKeyRequestCtx stored in context with value that *fasthttp.RequestCtx
	ContextKeyRequestCtx contextKey = iota
)

// RequestFunc may take information from an HTTP request and put it into a
// request context. In Servers, RequestFuncs are executed prior to invoking the
// endpoint. In Clients, RequestFuncs are executed after creating the request
// but prior to invoking the HTTP client.
type RequestFunc func(context.Context, *fasthttp.Request) context.Context

// ResponseFunc may take information from a request context and use it to
// manipulate a Response. In Servers, RequestFuncs are executed after invoking
// the endpoint but prior to writing a response. In Clients, RequestFuncs are
// executed after a request has been made, but prior to it being decoded.
type ResponseFunc func(context.Context, *fasthttp.Response) context.Context

// SetResponseHeader returns a ResponseFunc that sets the given header.
func SetResponseHeader(key, val string) ResponseFunc {
	return func(ctx context.Context, resp *fasthttp.Response) context.Context {
		resp.Header.Set(key, val)
		return ctx
	}
}

// SetRequestHeader returns a RequestFunc that sets the given header.
func SetRequestHeader(key, val string) RequestFunc {
	return func(ctx context.Context, req *fasthttp.Request) context.Context {
		req.Header.Set(key, val)
		return ctx
	}
}
