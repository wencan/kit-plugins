package fasthttp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	http_transport "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/endpoint"

	fasthttp_transport "github.com/wencan/kit-plugins/transport/fasthttp"
)

func TestClient(t *testing.T) {
	// test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	url, err := url.Parse(server.URL)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := fasthttp_transport.NewClient(
		http.MethodPost,
		url,
		fasthttp_transport.EncodeJSONRequest,
		fasthttp_transport.DecodeJSONResponse,
		newResponse,
		releaseResponse)
	var endpoint endpoint.Endpoint = client.Endpoint()

	request := &Request{
		Num: 10,
	}
	response, err := endpoint(context.Background(), request)
	if err != nil {
		t.Fatal(err)
		return
	}

	result := response.(*Response).Result
	if result != 100 {
		t.Fatalf("Want: 100, have: %d", result)
	}
}
