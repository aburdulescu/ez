package main

import (
	"fmt"
	"sync"

	"github.com/aburdulescu/ez/ezt"
)

type Value ezt.GetResponse

type KV struct {
	mu   sync.RWMutex
	data map[string]Value
}

func NewKV() *KV {
	return &KV{data: make(map[string]Value)}
}

func (kv *KV) Add(k string, ifile ezt.IFile, peer string) {
	kv.mu.Lock()
	v, ok := kv.data[k]
	value := Value{IFile: ifile}
	if ok {
		value.Peers = v.Peers
		if findInSlice(v.Peers, peer) == -1 {
			value.Peers = append(v.Peers, peer)
		}
	} else {
		value.Peers = []string{peer}
	}
	kv.data[k] = value
	kv.mu.Unlock()
}

func (kv *KV) Del(k string, peer string) error {
	kv.mu.Lock()
	v, ok := kv.data[k]
	if !ok {
		kv.mu.Unlock()
		return fmt.Errorf("key '%s' does not exist", k)
	}
	i := findInSlice(v.Peers, peer)
	if i == -1 {
		kv.mu.Unlock()
		return fmt.Errorf("peer '%s' does not exist for key '%s'", peer, k)
	}
	v.Peers[i] = v.Peers[len(v.Peers)-1]
	v.Peers = v.Peers[:len(v.Peers)-1]
	kv.data[k] = v
	if len(v.Peers) != 0 {
		kv.mu.Unlock()
		return nil
	}
	delete(kv.data, k)
	kv.mu.Unlock()
	return nil
}

func (kv *KV) Get(k string) (Value, error) {
	kv.mu.RLock()
	v, ok := kv.data[k]
	if !ok {
		return Value{}, fmt.Errorf("key '%s' does not exist", k)
	}
	kv.mu.RUnlock()
	return v, nil
}

func (kv *KV) List() []Value {
	kv.mu.RLock()
	values := make([]Value, len(kv.data))
	i := 0
	for _, v := range kv.data {
		values[i] = v
		i++
	}
	kv.mu.RUnlock()
	return values
}

func (kv *KV) GetAll() []ezt.GetAllItem {
	kv.mu.RLock()
	values := make([]ezt.GetAllItem, len(kv.data))
	i := 0
	for k, v := range kv.data {
		values[i] = ezt.GetAllItem{
			Hash: k,
			Name: v.IFile.Name,
			Size: v.IFile.Size,
		}
		i++
	}
	kv.mu.RUnlock()
	return values
}

func (kv *KV) Stat() uint64 {
	var stat uint64
	kv.mu.RLock()
	for k, v := range kv.data {
		stat += uint64(len(k))
		stat += uint64(len(v.IFile.Name) + 8 + 8)
		for i := range v.Peers {
			stat += uint64(len(v.Peers[i]))
		}
	}
	kv.mu.RUnlock()
	return stat
}

func findInSlice(s []string, v string) int {
	for i := range s {
		if s[i] == v {
			return i
		}
	}
	return -1
}
