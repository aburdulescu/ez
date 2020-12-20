package main

import (
	"log"
	"sync"
	"testing"
	"time"
)

type TestSeederClientDialer struct {
}

func TestConnPool(t *testing.T) {
	peers := []string{
		"1.1.1.1:111",
	}

	dialFunc := func(addr string) (*SeederClient, error) {
		return &SeederClient{}, nil
	}
	pool, err := NewConnPool(peers, dialFunc)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Release()

	peer := "1.1.1.1:111"

	log.Println(len(pool.data[peer]))

	n := 1000

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go testConnPoolGet(i, &wg, pool, peer)
	}
	wg.Wait()

	log.Println(len(pool.data[peer]))
}

func testConnPoolGet(i int, wg *sync.WaitGroup, pool *ConnPool, peer string) {
	defer wg.Done()
	client, err := pool.Get(peer)
	if err != nil {
		log.Println(err)
		return
	}
	defer pool.Put(peer, client)
	time.Sleep(1 * time.Second)
	if err := client.Connect("dummy"); err != nil {
		return
	}
}
