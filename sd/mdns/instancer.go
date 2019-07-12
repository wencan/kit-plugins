package mdns

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/hashicorp/mdns"

	"github.com/wencan/kit-plugins/sd/internal/instance"
)

const (
	defaultRefreshInterval = time.Second * 3
)

// InstancerOptions is used to customize how a Lookup is performed.
type InstancerOptions struct {
	RefreshInterval     time.Duration  // Refresh intervals, default 3 second
	Domain              string         // Lookup domain, default "local"
	LookupTimeout       time.Duration  // Lookup timeout, default 1 second
	Interface           *net.Interface // Multicast interface to use
	WantUnicastResponse bool           // Unicast response desired, as per 5.4 in RFC
}

// Instancer an mDns instancer. It will flushes the cache at intervals.
type Instancer struct {
	service string
	opts    InstancerOptions

	cache *instance.Cache

	logger log.Logger

	cancel func()
	wg     *sync.WaitGroup
}

// NewInstancer returns an mDNS instancer.
func NewInstancer(service string, opts InstancerOptions, logger log.Logger) (*Instancer, error) {
	if opts.RefreshInterval == 0 {
		opts.RefreshInterval = defaultRefreshInterval
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	inst := &Instancer{
		service: service,
		opts:    opts,
		cache:   instance.NewCache(),
		logger:  logger,
		cancel:  cancel,
		wg:      &wg,
	}

	// first lookup
	inst.refresh(ctx)

	go inst.loop(ctx)

	return inst, nil
}

func (inst *Instancer) loop(ctx context.Context) {
	inst.wg.Add(1)
	defer inst.wg.Done()

	refreshTicker := time.NewTicker(inst.opts.RefreshInterval)
	defer refreshTicker.Stop()

	for {
		select {
		case <-refreshTicker.C:
			inst.wg.Add(1)
			go func() {
				defer inst.wg.Done()
				inst.refresh(ctx)
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (inst *Instancer) refresh(ctx context.Context) {
	instances, err := inst.lookup(ctx)
	if err != nil {
		inst.cache.Update(sd.Event{Err: err})
	} else {
		inst.cache.Update(sd.Event{Instances: instances})
	}
}

// lookup looks up a given service, in a domain, waiting at most
// for a timeout before finishing the query.
func (inst *Instancer) lookup(ctx context.Context) ([]string, error) {
	entriesChan := make(chan *mdns.ServiceEntry, 100)
	instances := make([]string, 0)

	var lookupWG sync.WaitGroup
	lookupWG.Add(1)
	go func() {
		defer lookupWG.Done()

		for {
			select {
			case entry, ok := <-entriesChan:
				if !ok {
					return
				}
				instance, err := getInstance(entry)
				if err != nil {
					inst.logger.Log("action", "lookup", "err", err)
				} else {
					instances = append(instances, instance)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	param := &mdns.QueryParam{
		Service:             inst.service,
		Domain:              inst.opts.Domain,
		Timeout:             inst.opts.LookupTimeout,
		Interface:           inst.opts.Interface,
		Entries:             entriesChan,
		WantUnicastResponse: inst.opts.WantUnicastResponse,
	}
	err := mdns.Query(param)
	close(entriesChan)
	lookupWG.Wait()

	if err != nil {
		inst.logger.Log("action", "query", "err", err)
		return nil, err
	}
	return instances, nil
}

// getInstance get the instance address from mdns.ServiceEntry.
func getInstance(entry *mdns.ServiceEntry) (string, error) {
	if entry.AddrV4 != nil {
		instance := fmt.Sprintf("%s:%d", entry.AddrV4.String(), entry.Port)
		return instance, nil
	} else if entry.AddrV6 != nil {
		instance := fmt.Sprintf("%s:%d", entry.AddrV6.String(), entry.Port)
		return instance, nil
	} else {
		err := fmt.Errorf("invalid mdns entry: %v", entry)
		return "", err
	}
}

// Register implements Instancer.
func (inst *Instancer) Register(ch chan<- sd.Event) {
	inst.cache.Register(ch)
}

// Deregister implements Instancer.
func (inst *Instancer) Deregister(ch chan<- sd.Event) {
	inst.cache.Deregister(ch)
}

// State returns the current state of discovery (instances or error) as sd.Event
func (inst *Instancer) State() sd.Event {
	return inst.cache.State()
}

// Stop terminates the Instancer.
func (inst *Instancer) Stop() {
	inst.cancel()
	inst.cache.Stop()
	inst.wg.Wait()
}
