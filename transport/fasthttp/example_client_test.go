package fasthttp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	http_transport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/valyala/fasthttp"

	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		releaseRequest(request)
		return
	}

	response := newResponse()
	response.(*Response).Result = request.Num * request.Num
	err = http_transport.EncodeJSONResponse(context.Background(), w, response)
	releaseResponse(response)
	if err != nil {
		fmt.Println(err)
	}
}

func decodeResponse(ctx context.Context, resp *fasthttp.Response, response interface{}) error {
	// The fasthttp client always reads the whole body into memory before returning to the program.
	// https://github.com/valyala/fasthttp/issues/246
	return json.Unmarshal(resp.Body(), response)
}

func ExampleClient() {
	// test server
	server := httptest.NewServer(http.HandlerFunc(handler))
	url, err := url.Parse(server.URL)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := fasthttp_transport.NewClient(http.MethodPost, url, fasthttp_transport.EncodeJSONRequest, decodeResponse, newResponse, releaseResponse)
	var endpoint endpoint.Endpoint = client.Endpoint()

	request := &Request{
		Num: 10,
	}
	response, err := endpoint(context.Background(), request)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Result:", response.(*Response).Result)

	// Output:
	// Result: 100
}
