package protobuf

import (
	"bytes"
	"context"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/gogo/protobuf/proto/proto3_proto"
)

var (
	testMessage = proto3_proto.Message{
		Name: "Test",
	}
	testBuffer = []byte{10, 4, 84, 101, 115, 116}
)

func TestEncodeRequest(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	err := EncodeProtobufRequest(context.Background(), req, &testMessage)
	if err != nil {
		t.Fatal(err)
		return
	}

	contentType := req.Header.ContentType()
	if bytes.Compare(contentType, []byte(ProtobufContentType)) != 0 {
		t.Fatalf("Want: %s, have: %s", ProtobufContentType, string(contentType))
		return
	}
	contentLength := req.Header.ContentLength()
	if contentLength != len(testBuffer) {
		t.Fatalf("Want %d, have: %d", len(testBuffer), contentLength)
		return
	}
	if bytes.Compare(testBuffer, req.Body()) != 0 {
		t.Fatalf("Want %+v, have: %+v", testBuffer, req.Body())
		return
	}
}

func TestDecodeRequest(t *testing.T) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetBody(testBuffer)

	var message proto3_proto.Message
	err := DecodeProtobufRequest(context.Background(), req, &message)
	if err == nil {
		t.Fatalf("Want 'Content-Type not's %s', have: nil", ProtobufContentType)
		return
	}

	req.Header.SetContentType(ProtobufContentType)
	err = DecodeProtobufRequest(context.Background(), req, &message)
	if err != nil {
		t.Fatal(err)
		return
	}

	if testMessage.Name != message.Name {
		t.Fatalf("Want %+v, have: %+v", testMessage, message)
		return
	}
}

func TestEncodeResponse(t *testing.T) {
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := EncodeProtobufResponse(context.Background(), resp, &testMessage)
	if err != nil {
		t.Fatal(err)
		return
	}

	contentType := resp.Header.ContentType()
	if bytes.Compare(contentType, []byte(ProtobufContentType)) != 0 {
		t.Fatalf("Want: %s, have: %s", ProtobufContentType, string(contentType))
		return
	}
	contentLength := resp.Header.ContentLength()
	if contentLength != len(testBuffer) {
		t.Fatalf("Want %d, have: %d", len(testBuffer), contentLength)
		return
	}
	if bytes.Compare(testBuffer, resp.Body()) != 0 {
		t.Fatalf("Want %+v, have: %+v", testBuffer, resp.Body())
		return
	}
}

func TestDecodeResponse(t *testing.T) {
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	resp.SetBody(testBuffer)

	var message proto3_proto.Message
	err := DecodeProtobufResponse(context.Background(), resp, &message)
	if err == nil {
		t.Fatalf("Want 'Content-Type not's %s', have: nil", ProtobufContentType)
		return
	}

	resp.Header.SetContentType(ProtobufContentType)
	err = DecodeProtobufResponse(context.Background(), resp, &message)
	if err != nil {
		t.Fatal(err)
		return
	}

	if testMessage.Name != message.Name {
		t.Fatalf("Want %+v, have: %+v", testMessage, message)
		return
	}
}
