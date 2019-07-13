package mdns

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
)

func Example() {
	var (
		serverName = "/services/kit-mdns"
		instance   = "127.0.0.1:8080"
		port       = 8080

		logger = log.NewLogfmtLogger(os.Stdout)
	)

	// Build the registrar
	service := Service{
		Instance: instance,
		Service:  serverName,
		Port:     port,
		Ips:      []net.IP{net.IPv4(127, 0, 0, 1)}, // Just for test
	}
	registrar, err := NewRegistrar(service, logger)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Register my instance
	registrar.Register()
	defer registrar.Deregister()

	// Build the instancer
	instancer, err := NewInstancer(serverName, InstancerOptions{}, logger)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Build the endpoint
	endpointer := sd.NewEndpointer(instancer, fakeFactory, logger)
	_, err = endpointer.Endpoints()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// 127.0.0.1:8080
}

func fakeFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	// Print instance
	fmt.Println(instance)

	return endpoint.Nop, ioutil.NopCloser(nil), nil
}
