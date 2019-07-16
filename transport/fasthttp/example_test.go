package fasthttp_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/valyala/fasthttp"
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

func newServerEndpoint(newResponse fasthttp_transport.NewObjectFunc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r := request.(*Request)
		res := newResponse().(*Response)
		res.Result = r.Num * r.Num
		return res, nil
	}
}

func Example() {
	// Create a server
	server := fasthttp_transport.NewServer(
		newServerEndpoint(newResponse),
		fasthttp_transport.DecodeJSONRequest,
		fasthttp_transport.EncodeJSONResponse,
		newRequest,
		releaseRequest,
		releaseResponse)

	// Listen a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run a server unit shutdown
	s := fasthttp.Server{
		Handler: server.ServeFastHTTP,
	}
	go func() {
		if err := s.Serve(listener); err != nil {
			fmt.Printf("Error in server: %s", err)
		}
	}()
	defer s.Shutdown()

	// Create a client
	url, err := url.Parse(fmt.Sprintf("http://%s/", listener.Addr().String()))
	if err != nil {
		fmt.Println(err)
		return
	}
	opt := fasthttp_transport.SetClient(&fasthttp.Client{
		MaxIdleConnDuration: time.Millisecond * 10, // Just for test
	})
	client := fasthttp_transport.NewClient(
		http.MethodPost,
		url,
		fasthttp_transport.EncodeJSONRequest,
		fasthttp_transport.DecodeJSONResponse,
		newResponse,
		releaseResponse,
		opt)

	// Create endpoint
	endpoint := client.Endpoint()

	// Call
	request := &Request{
		Num: 10,
	}
	response, err := endpoint(context.Background(), request)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Result: %d", response.(*Response).Result)

	// Output:
	// Result: 100
}
