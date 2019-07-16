package fasthttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/valyala/fasthttp"
)

// ErrorEncoder is responsible for encoding an error to the ResponseWriter.
// Users are encouraged to use custom ErrorEncoders to encode HTTP errors to
// their clients, and will likely want to pass and check for their own error
// types. See the example shipping/handling service.
type ErrorEncoder func(ctx context.Context, err error, resp *fasthttp.Response)

// ServerFinalizerFunc can be used to perform work at the end of an HTTP
// request, after the response has been written to the client. The principal
// intended use is for request logging.
type ServerFinalizerFunc func(c context.Context, req *fasthttp.Request, resp *fasthttp.Response, err error)

// NopReleaser is an ReleaseObjectFunc that do nothing.
// It's designed to be used in compatible servers.
func NopReleaser(_ interface{}) {}

// Server wraps an endpoint and provide fasthttp.RequestHandler method.
type Server struct {
	e               endpoint.Endpoint
	dec             DecodeRequestFunc
	enc             EncodeResponseFunc
	newRequest      NewObjectFunc
	releaseRequest  ReleaseObjectFunc
	releaseResponse ReleaseObjectFunc
	before          []RequestFunc
	after           []ResponseFunc
	errorEncoder    ErrorEncoder
	finalizer       []ServerFinalizerFunc
	errorHandler    transport.ErrorHandler
}

// NewServer constructs a new server.
func NewServer(e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	newRequest NewObjectFunc,
	releaseRequest ReleaseObjectFunc,
	releaseResponse ReleaseObjectFunc,
	options ...ServerOption) *Server {
	s := &Server{
		e:               e,
		dec:             dec,
		enc:             enc,
		newRequest:      newRequest,
		releaseRequest:  releaseRequest,
		releaseResponse: releaseResponse,
		errorEncoder:    DefaultErrorEncoder,
		errorHandler:    transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// NewCompatibleServer constructs a new compatible server.
// It does not reuse response object to share endpoint with other transport servers.
func NewCompatibleServer(e endpoint.Endpoint,
	dec DecodeRequestFunc,
	enc EncodeResponseFunc,
	newRequest NewObjectFunc,
	releaseRequest ReleaseObjectFunc,
	options ...ServerOption) *Server {
	return NewServer(e, dec, enc, newRequest, releaseRequest, NopReleaser, options...)
}

// ServerOption sets an optional parameter for servers.
type ServerOption func(*Server)

// ServerBefore functions are executed on the HTTP request object before the
// request is decoded.
func ServerBefore(before ...RequestFunc) ServerOption {
	return func(s *Server) { s.before = append(s.before, before...) }
}

// ServerAfter functions are executed on the HTTP response writer after the
// endpoint is invoked, but before anything is written to the client.
func ServerAfter(after ...ResponseFunc) ServerOption {
	return func(s *Server) { s.after = append(s.after, after...) }
}

// ServerErrorEncoder is used to encode errors to the http.ResponseWriter
// whenever they're encountered in the processing of a request. Clients can
// use this to provide custom error formatting and response codes. By default,
// errors will be written with the DefaultErrorEncoder.
func ServerErrorEncoder(ee ErrorEncoder) ServerOption {
	return func(s *Server) { s.errorEncoder = ee }
}

// ServerErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom ServerErrorEncoder or ServerFinalizer, both of which have access to
// the context.
func ServerErrorHandler(errorHandler transport.ErrorHandler) ServerOption {
	return func(s *Server) { s.errorHandler = errorHandler }
}

// ServerFinalizer is executed at the end of every HTTP request.
// By default, no finalizer is registered.
func ServerFinalizer(f ...ServerFinalizerFunc) ServerOption {
	return func(s *Server) { s.finalizer = append(s.finalizer, f...) }
}

// ServeFastHTTP provide fasthttp.RequestHandler method.
func (s Server) ServeFastHTTP(ctx *fasthttp.RequestCtx) {
	c := context.WithValue(context.Background(), ContextKeyRequestCtx, ctx)
	var err error

	if len(s.finalizer) > 0 {
		defer func() {
			for _, f := range s.finalizer {
				f(c, &ctx.Request, &ctx.Response, err)
			}
		}()
	}

	for _, f := range s.before {
		c = f(c, &ctx.Request)
	}

	request := s.newRequest()

	err = s.dec(c, &ctx.Request, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, &ctx.Response)
		s.releaseRequest(request)
		return
	}

	var response interface{}
	response, err = s.e(c, request)
	s.releaseRequest(request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, &ctx.Response)
		return
	}

	for _, f := range s.after {
		c = f(c, &ctx.Response)
	}

	err = s.enc(c, &ctx.Response, response)
	s.releaseResponse(response)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		s.errorEncoder(ctx, err, &ctx.Response)
		return
	}
}

// DefaultErrorEncoder writes the error to the ResponseWriter, by default a
// content type of text/plain, a body of the plain text of the error, and a
// status code of 500. If the error implements Headerer, the provided headers
// will be applied to the response. If the error implements json.Marshaler, and
// the marshaling succeeds, a content type of application/json and the JSON
// encoded form of the error will be used. If the error implements StatusCoder,
// the provided StatusCode will be used instead of 500.
func DefaultErrorEncoder(_ context.Context, err error, resp *fasthttp.Response) {
	contentType, body := "text/plain; charset=utf-8", []byte(err.Error())
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			contentType, body = "application/json; charset=utf-8", jsonBody
		}
	}
	resp.Header.SetContentType(contentType)
	if headerer, ok := err.(Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				resp.Header.Add(k, v)
			}
		}
	}
	code := http.StatusInternalServerError
	if sc, ok := err.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	resp.SetStatusCode(code)
	resp.SetBody(body)
}

// StatusCoder is checked by DefaultErrorEncoder. If an error value implements
// StatusCoder, the StatusCode will be used when encoding the error. By default,
// StatusInternalServerError (500) is used.
type StatusCoder interface {
	StatusCode() int
}

// Headerer is checked by DefaultErrorEncoder. If an error value implements
// Headerer, the provided headers will be applied to the response writer, after
// the Content-Type is set.
type Headerer interface {
	Headers() http.Header
}
