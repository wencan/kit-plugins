package mdns

import (
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/mdns"
)

func newTestRegistrar(serviceName string, port int) (*Registrar, string, error) {
	ips := []net.IP{net.IPv4(127, 0, 0, 1)}
	instance := fmt.Sprintf("%s:%d", ips[0].String(), port)

	service := Service{
		Instance: instance,
		Service:  serviceName,
		Ips:      ips, // Just for test
		Port:     port,
	}
	registrar, err := NewRegistrar(service, log.NewNopLogger())
	return registrar, instance, err
}

func TestRegistrar(t *testing.T) {
	serviceName := "test.registrar.mdns.kit"

	// Create mDNS registrar
	registrars := []*Registrar{}
	defer func() {
		for _, registrar := range registrars {
			registrar.Deregister()
		}
	}()
	want := []string{}
	port := 0
	for i := 0; i < 10; i++ {
		port += rand.Intn(1000) + 1
		registrar, instance, err := newTestRegistrar(serviceName, port)
		if err != nil {
			t.Fatal(err)
		}
		registrar.Register()

		registrars = append(registrars, registrar)
		want = append(want, instance)
	}

	// Create the mDNS instancer
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	have := []string{}
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for entry := range entriesCh {
			if entry.AddrV4 == nil {
				err = fmt.Errorf("entry %s no IPv4", entry.Name)
			} else {
				instance := fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)
				have = append(have, instance)
			}
		}
	}()

	// Lookup instances
	mdns.Lookup(serviceName, entriesCh)
	close(entriesCh)
	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(want)
	sort.Strings(have)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want: %s have: %s", want, have)
	}
}
