package fasthttp

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/go-kit/kit/endpoint"
	"github.com/valyala/fasthttp"
)

// ClientFinalizerFunc can be used to perform work at the end of a client HTTP
// request, after the response is returned. The principal
// intended use is for error logging.
// Note: err may be nil. There maybe also no additional response parameters
// depending on when an error occurs.
type ClientFinalizerFunc func(c context.Context, req *fasthttp.Request, resp *fasthttp.Response, err error)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
type Client struct {
	method          string
	tgt             *url.URL
	enc             EncodeRequestFunc
	dec             DecodeResponseFunc
	newResponse     NewObjectFunc
	releaseResponse ReleaseObjectFunc
	before          []RequestFunc
	after           []ClientResponseFunc
	finalizer       []ClientFinalizerFunc
}

// NewClient constructs a usable Client for a single remote method.
func NewClient(
	method string,
	tgt *url.URL,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	newResponse NewObjectFunc,
	releaseResponse ReleaseObjectFunc,
	options ...ClientOption,
) *Client {
	client := &Client{
		method:          method,
		tgt:             tgt,
		enc:             enc,
		dec:             dec,
		newResponse:     newResponse,
		releaseResponse: releaseResponse,
		before:          []RequestFunc{},
		after:           []ClientResponseFunc{},
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// ClientBefore sets the RequestFuncs that are applied to the outgoing HTTP
// request before it's invoked.
func ClientBefore(before ...RequestFunc) ClientOption {
	return func(client *Client) { client.before = append(client.before, before...) }
}

// ClientAfter sets the ClientResponseFuncs applied to the incoming HTTP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(client *Client) { client.after = append(client.after, after...) }
}

// ClientFinalizer is executed at the end of every HTTP request.
// By default, no finalizer is registered.
func ClientFinalizer(f ...ClientFinalizerFunc) ClientOption {
	return func(s *Client) { s.finalizer = append(s.finalizer, f...) }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (client Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		c := context.Background()

		var (
			req  *fasthttp.Request
			resp *fasthttp.Response
			err  error
		)
		if client.finalizer != nil {
			defer func() {
				for _, f := range client.finalizer {
					f(c, req, resp, err)
				}
			}()
		}

		req = fasthttp.AcquireRequest()
		req.Header.SetMethod(client.method)
		req.SetRequestURI(client.tgt.String())

		if err = client.enc(c, req, request); err != nil {
			fasthttp.ReleaseRequest(req)
			return nil, err
		}

		for _, f := range client.before {
			c = f(c, req)
		}

		resp = fasthttp.AcquireResponse()
		err = fasthttp.Do(req, resp)
		fasthttp.ReleaseRequest(req)
		if err != nil {
			fasthttp.ReleaseResponse(resp)
			return nil, err
		}

		for _, f := range client.after {
			ctx = f(ctx, resp)
		}

		response := client.newResponse()
		err = client.dec(c, resp, response)
		fasthttp.ReleaseResponse(resp)
		if err != nil {
			client.releaseResponse(response)
			return nil, err
		}

		return response, nil
	}
}

// EncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Request body. Many JSON-over-HTTP services can use it as
// a sensible default. If the request implements Headerer, the provided headers
// will be applied to the request.
func EncodeJSONRequest(ctx context.Context, r *fasthttp.Request, request interface{}) error {
	r.Header.SetContentType("application/json; charset=utf-8")
	if headerer, ok := request.(Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	return json.NewEncoder(r.BodyWriter()).Encode(request)
}
