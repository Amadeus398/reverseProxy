package backendManager

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/backends"
	"sync"
	"time"
)

var (
	BackendMgr        *BackendManager
	ErrNoHost         = fmt.Errorf("host not found")
	ErrClientNotFound = fmt.Errorf("client not found")
)

type BackendManager struct {
	endPoints   map[string][]*Client
	tickBackend *time.Ticker
	tickDB      *time.Ticker
	ctx         context.Context
	mux         sync.RWMutex
	e           chan error
	log         *logging.Logger
}

type Client struct {
	Alive     bool
	Address   string
	processed bool
	Cl        http.Client
	mux       sync.RWMutex
}

// NewBackendManager returns new struct BackendManager
func NewBackendManager(ctx context.Context) *BackendManager {
	return &BackendManager{
		endPoints:   make(map[string][]*Client),
		tickBackend: time.NewTicker(20 * time.Second),
		tickDB:      time.NewTicker(5 * time.Second),
		ctx:         ctx,
		e:           make(chan error),
	}
}

// syncHosts updates the current hosts and clients of endpoints
func (b *BackendManager) syncHosts(endpoints []*backends.Backend) error {
	b.log = logging.NewLogs("backendManager", "syncHosts")

	b.log.GetInfo().Msg("updating the host data")
	for host, val := range b.endPoints {
		match := false
		for _, endpoint := range endpoints {
			if endpoint.Site.Host == host {
				match = true
				break
			}
		}
		if !match {
			for _, client := range val {
				client.Cl.CloseIdleConnections()
			}
			delete(b.endPoints, host)
		} else {
			for _, client := range val {
				client.processed = false
			}
		}
	}

	b.log.GetInfo().Msg("updating clients by host")
	for _, endpoint := range endpoints {
		match := false
		_, ok := b.endPoints[endpoint.Site.Host]
		if !ok {
			b.endPoints[endpoint.Site.Host] = []*Client{}
		}
		for _, client := range b.endPoints[endpoint.Site.Host] {
			if client.Address == endpoint.Address {
				client.processed = true
				match = true
				break
			}
		}
		if !match {
			client := Client{
				Address:   endpoint.Address,
				processed: true,
			}
			b.endPoints[endpoint.Site.Host] = append(b.endPoints[endpoint.Site.Host], &client)
		}
	}

	b.log.GetInfo().Msg("adding new clients by host")
	for key, clients := range b.endPoints {
		newClients := []*Client{}
		for _, client := range clients {
			if client.processed {
				newClients = append(newClients, client)
			} else {
				client.Cl.CloseIdleConnections()
			}
		}
		b.endPoints[key] = newClients
	}
	return nil
}

// SyncEndpoints updates the current endpoints of database
func (b *BackendManager) SyncEndpoints() {
	b.mux.Lock()
	defer b.mux.Unlock()
	endpoint, err := backends.List()
	if err != nil {
		b.e <- err
		return
	}
	if err := b.syncHosts(endpoint); err != nil {
		b.e <- err
	}
}

// CheckEndpoints check Alive clients
func (b *BackendManager) CheckEndpoints() {
	b.mux.RLock()
	defer b.mux.RUnlock()

	for _, clients := range b.endPoints {
		for _, client := range clients {
			go client.ping()
		}
	}
}

// ping establishes a connection with each
// client and updates the status Alive
func (c *Client) ping() {
	c.mux.Lock()
	defer c.mux.Unlock()
	conn, err := net.DialTimeout("tcp", c.Address, 1*time.Second)
	if err != nil {
		c.Alive = false
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			return
		}
	}()
	c.Alive = true
}

// getAlive block the safe reading of the
// client's status Alive
func (c *Client) getAlive() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.Alive
}

// GetClient randomly selects a client for
// a given host
func (b *BackendManager) GetClient(host string) (*Client, error) {
	b.log = logging.NewLogs("backendManager", "getClient")

	b.mux.RLock()
	defer b.mux.RUnlock()

	clients, ok := b.endPoints[host]
	if !ok {
		b.log.GetWarn().Msg("host not found")
		return nil, ErrNoHost
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(clients))))
	if err != nil {
		b.log.GetError().Str("when", "get random cryptographically integer").
			Err(err).Msg("failed get integer")
		return nil, err
	}
	n := nBig.Int64()
	client := clients[n]
	if !client.getAlive() {
		b.log.GetWarn().Msg("client alive false")
		i := n + 1
		for i != n {
			if i == int64(len(clients)) {
				i = 0
			}
			if clients[i].getAlive() {
				b.log.GetWarn().Msg("client alive true")
				return clients[i], nil
			}
		}
	} else {
		return client, nil
	}

	b.log.GetWarn().Msg("client not found")
	return nil, ErrClientNotFound
}

// Serve with the ticks running SyncEndpoints
// and CheckEndpoints during the operation of
// the application
func (b *BackendManager) Serve() error {
	defer b.tickDB.Stop()
	defer b.tickBackend.Stop()
	defer close(b.e)
	for {
		select {
		case <-b.tickDB.C:
			go b.SyncEndpoints()
		case <-b.tickBackend.C:
			go b.CheckEndpoints()
		case err := <-b.e:
			return err
		case <-b.ctx.Done():
			return nil
		}
	}
}
