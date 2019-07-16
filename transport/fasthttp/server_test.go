package fasthttp_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

func TestServer(t *testing.T) {
	// Kit server
	server := fasthttp_transport.NewServer(
		newServerEndpoint(newResponse),
		fasthttp_transport.DecodeJSONRequest,
		fasthttp_transport.EncodeJSONResponse,
		newRequest,
		releaseRequest,
		releaseResponse)

	// Mock listener
	listener := fasthttputil.NewInmemoryListener()
	defer listener.Close()
	// serve
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		fasthttp.ServeConn(conn, server.ServeFastHTTP)
	}()

	// Mock client
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return listener.Dial()
			},
		},
	}

	// Post
	res, err := client.Post("http://test/", "application/json; charset=utf-8", bytes.NewBufferString("{\"Num\": 10}"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Check response
	if res.Body == nil {
		panic("res.Body is nil")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
		return
	}
	sbody := strings.TrimSpace(string(body)) // remove \n
	if sbody != "{\"Result\":100}" {
		t.Fatalf("Want: {\"Result\":100}, have: %s", sbody)
	}
}
