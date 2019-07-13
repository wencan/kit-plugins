package mdns

import (
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hashicorp/mdns"
)

func newTestServer(serviceName string, port int) (*mdns.Server, string, error) {
	ips := []net.IP{net.IPv4(127, 0, 0, 1)} // Just for test
	instance := fmt.Sprintf("%s:%d", ips[0].String(), port)

	service, err := mdns.NewMDNSService(instance, serviceName, "", "", port, ips, nil)
	if err != nil {
		return nil, "", err
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	return server, instance, nil
}

func TestMDNSInstancer(t *testing.T) {
	serviceName := "test.instancer.mdns.kit"

	rand.Seed(time.Now().Unix())

	// Create mDNS servers
	servers := []*mdns.Server{}
	defer func() {
		for _, server := range servers {
			server.Shutdown()
		}
	}()
	want := []string{}
	port := 0
	for i := 0; i < 10; i++ {
		port += rand.Intn(1000) + 1
		server, instance, err := newTestServer(serviceName, port)
		if err != nil {
			t.Fatal(err)
		}
		servers = append(servers, server)
		want = append(want, instance)
	}

	// Create the mDNS instancer
	instancer, err := NewInstancer(serviceName, InstancerOptions{
		RefreshInterval: time.Second * 3,
	}, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer instancer.Stop()

	// Get the state of discovery (instances or error)
	event := instancer.State()
	if event.Err != nil {
		t.Fatal(event.Err)
	}
	have := event.Instances

	sort.Strings(want)
	sort.Strings(have)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want: %s have: %s", want, have)
	}
}
