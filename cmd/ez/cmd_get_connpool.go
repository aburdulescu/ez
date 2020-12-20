package main

import (
	"fmt"
	"log"
	"sync"
)

type ConnPoolDialFunc func(addr string) (*SeederClient, error)

type ConnPool struct {
	mu   sync.RWMutex
	data map[string][]*SeederClient

	dialFunc ConnPoolDialFunc
}

func NewConnPool(peers []string, dialFunc ConnPoolDialFunc) (*ConnPool, []string, error) {
	var goodPeers []string
	data := make(map[string][]*SeederClient)
	for _, peer := range peers {
		c, err := dialFunc(peer)
		if err != nil {
			log.Println(err)
			continue
		}
		data[peer] = append(data[peer], c)
		goodPeers = append(goodPeers, peer)
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no peers available")
	}
	pool := &ConnPool{
		data:     data,
		dialFunc: dialFunc,
	}
	return pool, goodPeers, nil
}

func (p *ConnPool) Get(id string, addr string) (*SeederClient, error) {
	p.mu.Lock()
	clients, ok := p.data[addr]
	if ok && len(clients) != 0 {
		client := clients[0]
		clients = clients[1:]
		if len(clients) == 0 {
			delete(p.data, addr)
		} else {
			p.data[addr] = clients
		}
		p.mu.Unlock()
		return client, nil
	} else {
		p.mu.Unlock()
		client, err := p.dialFunc(addr)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if err := client.Connect(id); err != nil {
			client.Close()
			return nil, err
		}
		return client, nil
	}
}

func (p *ConnPool) Put(addr string, client *SeederClient) {
	p.mu.Lock()
	p.data[addr] = append(p.data[addr], client)
	p.mu.Unlock()
}

func (p *ConnPool) Len() int {
	p.mu.RLock()
	l := len(p.data)
	p.mu.RUnlock()
	return l
}

func (p *ConnPool) Connect(id string) error {
	p.mu.RLock()
	for _, clients := range p.data {
		for _, client := range clients {
			if err := client.Connect(id); err != nil {
				// TODO: remove conn if connect fails
				return err
			}
		}
	}
	p.mu.RUnlock()
	return nil
}

func (p *ConnPool) Disconnect() {
	p.mu.RLock()
	for _, clients := range p.data {
		for _, client := range clients {
			client.Disconnect()
		}
	}
	p.mu.RUnlock()
}

func (p *ConnPool) Release() {
	p.mu.Lock()
	for _, clients := range p.data {
		for i := range clients {
			clients[i].Close()
		}
	}
	p.data = make(map[string][]*SeederClient)
	p.mu.Unlock()
}
