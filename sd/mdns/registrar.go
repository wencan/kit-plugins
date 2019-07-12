package mdns

import (
	"net"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/mdns"
)

// Service holds the instance config.
type Service struct {
	Instance string // Required and unique
	Service  string // Required
	Domain   string
	HostName string
	Port     int // Required
	Ips      []net.IP
	Txt      []string
}

// Registrar is used to listen for mDNS queries and respond if we have a matching local record.
type Registrar struct {
	config *mdns.Config
	server *mdns.Server

	logger log.Logger
}

// NewRegistrar is used to create a new registrar from a service config.
func NewRegistrar(service Service, logger log.Logger) (*Registrar, error) {
	zone, err := mdns.NewMDNSService(service.Instance, service.Service,
		service.Domain, service.HostName, service.Port, service.Ips, service.Txt)
	if err != nil {
		return nil, err
	}

	config := &mdns.Config{
		Zone: zone,
	}
	registrar := &Registrar{
		config: config,
		logger: logger,
	}
	return registrar, nil
}

// Register is used to listen for mDNS queries.
func (registrar *Registrar) Register() {
	if registrar.server != nil {
		registrar.logger.Log("action", "register", "err", "already registered")
		return
	}

	server, err := mdns.NewServer(registrar.config)
	if err != nil {
		registrar.logger.Log("action", "register", "err", err)
		return
	}

	registrar.server = server
}

// Deregister is used to shutdown the listener.
func (registrar *Registrar) Deregister() {
	if registrar.server == nil {
		registrar.logger.Log("action", "deregister", "err", "not registered")
		return
	}

	err := registrar.server.Shutdown()
	if err != nil {
		registrar.logger.Log("action", "deregister", "err", err)
		return
	}

	registrar.server = nil
}
