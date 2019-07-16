package fasthttp_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/fasthttp/router"
	"github.com/go-kit/kit/endpoint"
	"github.com/valyala/fasthttp"
	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

type HelloRequest struct {
	Name string
}

type HelloResponse struct {
	Greeting string
}

var (
	helloRequestPool = &sync.Pool{
		New: func() interface{} { return new(HelloRequest) },
	}
	helloResponsePool = &sync.Pool{
		New: func() interface{} { return new(HelloResponse) },
	}

	newHelloRequest = helloRequestPool.Get

	newHelloResponse = helloResponsePool.Get
)

func releaseHelloRequest(request interface{}) {
	r := request.(*HelloRequest)
	r.Name = ""             // clear
	helloRequestPool.Put(r) // release
}

func releaseHelloResponse(response interface{}) {
	res := response.(*HelloResponse)
	res.Greeting = ""          // clear
	helloResponsePool.Put(res) // release
}

func encodeHelloRequest(c context.Context, r *fasthttp.Request, request interface{}) error {
	ctx := c.Value(fasthttp_transport.ContextKeyRequestCtx).(*fasthttp.RequestCtx)
	name := ctx.UserValue("name").(string)

	request.(*HelloRequest).Name = name
	return nil
}

func newServerHelloEndpoint(newResponse fasthttp_transport.NewObjectFunc) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		r := request.(*HelloRequest)
		res := newResponse().(*HelloResponse)
		res.Greeting = fmt.Sprintf("hello, %s", r.Name)
		return res, nil
	}
}

func ExampleServer_router() {
	// Create hello server
	helloServer := fasthttp_transport.NewServer(
		newServerHelloEndpoint(newHelloResponse),
		encodeHelloRequest,
		newHelloRequest,
		releaseHelloRequest,
		fasthttp_transport.EncodeJSONResponse,
		releaseHelloResponse)

	// Create router
	router := router.New()
	router.GET("/hello/:name", helloServer.ServeFastHTTP)

	// Run fasthttp server
	err := fasthttp.ListenAndServe("127.0.0.1:8080", router.Handler)
	if err != nil {
		fmt.Println(err)
		return
	}
}
