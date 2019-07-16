package protobuf

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/valyala/fasthttp"
	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

var (
	// IngoreContentType Content-Type header will be ignored when decoding the body
	IngoreContentType bool

	// ProtobufContentType Content-Type header value that indicates that the content is protocol buffer
	ProtobufContentType = "application/x-protobuf"

	// DefaultBufferSize Buffer default size
	DefaultBufferSize = 64

	bufferPool = sync.Pool{}
)

func acquireProtoBuffer() *proto.Buffer {
	buffer := bufferPool.Get()
	if buffer == nil {
		return proto.NewBuffer(make([]byte, DefaultBufferSize))
	}
	return buffer.(*proto.Buffer)
}

func releaseProtoBuffer(buffer *proto.Buffer) {
	buffer.Reset()
	bufferPool.Put(buffer)
}

// EncodeProtobufRequest is an fasthttp_transport.EncodeRequestFunc that serializes
// the request as a protobuf message object to the Request body. Many protobuf-over-HTTP
// services can use it as a sensible default. If the request implements Headerer,
// the provided headers will be applied to the request.
func EncodeProtobufRequest(_ context.Context, r *fasthttp.Request, request interface{}) error {
	r.Header.SetContentType(ProtobufContentType)
	if headerer, ok := request.(fasthttp_transport.Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}

	msg, ok := request.(proto.Message)
	if !ok {
		return errors.New("request does not implement proto.Message")
	}

	buffer := acquireProtoBuffer()
	err := buffer.Marshal(msg)
	if err != nil {
		releaseProtoBuffer(buffer)
		return err
	}

	r.Header.SetContentLength(len(buffer.Bytes()))
	r.SetBody(buffer.Bytes())
	releaseProtoBuffer(buffer)
	return nil
}

// DecodeProtobufRequest is an fasthttp_transport.DecodeRequestFunc that deserializes the
// response as a protobuf message object from the Request body. Many protobuf-over-HTTP
// services can use it as a sensible default.
func DecodeProtobufRequest(_ context.Context, r *fasthttp.Request, request interface{}) error {
	if !IngoreContentType {
		contentType := strings.Split(b2s(r.Header.ContentType()), ";")[0]

		if contentType != ProtobufContentType {
			return fmt.Errorf("Content-Type not's %s", ProtobufContentType)
		}
	}

	msg, ok := request.(proto.Message)
	if !ok {
		return errors.New("request does not implement proto.Message")
	}

	return proto.Unmarshal(r.Body(), msg)
}

// EncodeProtobufResponse is a fasthttp_transport.EncodeResponseFunc that serializes
// the response as a protobuf message object to the Response. Many protobuf-over-HTTP
// services can use it as a sensible default. If the response implements Headerer,
// the provided headers will be applied to the response. If the response implements
// StatusCoder, the provided StatusCode will be used instead of 200.
func EncodeProtobufResponse(_ context.Context, resp *fasthttp.Response, response interface{}) error {
	resp.Header.SetContentType(ProtobufContentType)
	if headerer, ok := response.(fasthttp_transport.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				resp.Header.Add(k, v)
			}
		}
	}

	code := http.StatusOK
	if sc, ok := response.(fasthttp_transport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	resp.SetStatusCode(code)
	if code == http.StatusNoContent {
		return nil
	}

	msg, ok := response.(proto.Message)
	if !ok {
		return errors.New("request does not implement proto.Message")
	}

	buffer := acquireProtoBuffer()
	err := buffer.Marshal(msg)
	if err != nil {
		releaseProtoBuffer(buffer)
		return err
	}

	resp.Header.SetContentLength(len(buffer.Bytes()))
	resp.SetBody(buffer.Bytes())
	releaseProtoBuffer(buffer)
	return nil
}

// DecodeProtobufResponse is an fasthttp_transport.DecodeResponseFunc that deserializes
// the response as a response object from the Response body. Many protobuf-over-HTTP
// services can use it as a sensible default.
func DecodeProtobufResponse(_ context.Context, resp *fasthttp.Response, response interface{}) error {
	if !IngoreContentType {
		contentType := strings.Split(b2s(resp.Header.ContentType()), ";")[0]

		if contentType != ProtobufContentType {
			return fmt.Errorf("Content-Type not's %s", ProtobufContentType)
		}
	}

	msg, ok := response.(proto.Message)
	if !ok {
		return errors.New("request does not implement proto.Message")
	}

	// The fasthttp client always reads the whole body into memory before returning to the program.
	// https://github.com/valyala/fasthttp/issues/246
	return proto.Unmarshal(resp.Body(), msg)
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
