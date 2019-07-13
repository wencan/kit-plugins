package fasthttp_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

type Request struct {
	Num int
}

type Response struct {
	Result int
}

var (
	requestPool = &sync.Pool{
		New: func() interface{} { return new(Request) },
	}
	responsePool = &sync.Pool{
		New: func() interface{} { return new(Response) },
	}

	newRequest = requestPool.Get

	newResponse = responsePool.Get
)

func releaseRequest(request interface{}) {
	r := request.(*Request)
	r.Num = 0          // clear
	requestPool.Put(r) // release
}

func releaseResponse(response interface{}) {
	res := response.(*Response)
	res.Result = 0        // clear
	responsePool.Put(res) // release
}

func newEndpoint(newResponse fasthttp_transport.NewObjectFunc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r := request.(*Request)
		res := newResponse().(*Response)
		res.Result = r.Num * r.Num
		return res, nil
	}
}

func encodeRequest(ctx context.Context, req *fasthttp.Request, request interface{}) (err error) {
	r := request.(*Request)
	r.Num, err = req.PostArgs().GetUint("num")
	return
}

func ExampleServer() {
	// Kit server
	server := fasthttp_transport.NewServer(newEndpoint(newResponse), encodeRequest, newRequest, releaseRequest, fasthttp_transport.EncodeJSONResponse, releaseResponse)

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
	res, err := client.PostForm("http://test/", url.Values(map[string][]string{"num": []string{"10"}}))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check response
	if res.Body == nil {
		panic("res.Body is nil")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

	// Output:
	// {"Result":100}
}
