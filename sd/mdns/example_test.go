package mdns

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

func Example() {
	var (
		serverName = "/services/kit-mdns"
		instance   = "127.0.0.1:8080"
		port       = 8080

		logger = log.NewLogfmtLogger(&bytes.Buffer{})
	)

	// Build the registrar
	service := Service{
		Instance: instance,
		Service:  serverName,
		Port:     port,
	}
	registrar, err := NewRegistrar(service, logger)
	if err != nil {
		logger.Log(err)
		return
	}
	// Register my instance
	registrar.Register()
	defer registrar.Deregister()

	// Build the instancer
	instancer, err := NewInstancer(serverName, InstancerOptions{}, logger)
	if err != nil {
		logger.Log(err)
		return
	}

	// Build the endpoint
	endpointer := sd.NewEndpointer(instancer, fakeFactory, logger)
	_ = endpointer
}

func fakeFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	return endpoint.Nop, ioutil.NopCloser(nil), nil
}
