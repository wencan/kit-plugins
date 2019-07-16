[![GoDoc](https://godoc.org/github.com/wencan/kit-plugins/sd/mdns?status.svg)](https://godoc.org/github.com/wencan/kit-plugins/sd/mdns)

# mdns
Package mdns provides Instancer and Registrar implementations for mDNS.

mDNS or Multicast DNS can be used to discover services on the local network without the use of an authoritative DNS server. This enables peer-to-peer discovery. It is important to note that many networks restrict the use of multicasting, which prevents mDNS from functioning. Notably, multicast cannot be used in any sort of cloud, or shared infrastructure environment. However it works well in most office, home, or private infrastructure environments.

# example
```go
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
```