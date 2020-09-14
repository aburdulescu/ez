package main

import (
	"log"
	"sync"
	"testing"

	"github.com/aburdulescu/go-ez/ezt"
)

func TestAdd(t *testing.T) {
	kv := NewKV()
	var wg sync.WaitGroup
	c := 1000
	wg.Add(c)
	for i := 0; i < c; i++ {
		go add(kv, &wg)
	}
	wg.Wait()
	t.Log("kv aprox. size:", kv.Stat())
}

func add(kv *KV, wg *sync.WaitGroup) {
	defer wg.Done()

	input := map[string]Value{
		"1": {ezt.IFile{"a", "/a", 1}, []string{"1", "2", "3"}},
		"2": {ezt.IFile{"b", "/b", 2}, []string{"2", "3"}},
		"3": {ezt.IFile{"c", "/c", 3}, []string{"3"}},
	}

	expected := make(Values, len(input))
	n := 0
	for _, v := range input {
		expected[n] = v
		n++
	}

	for k, v := range input {
		for i := range v.Peers {
			kv.Add(k, v.IFile, v.Peers[i])
		}
	}

	values := kv.List()

	result := Values(values)
	if !result.equals(expected) {
		log.Println("expected:", expected)
		log.Println("result:", result)
		log.Fatal("result != expected")
	}
}

func TestDel(t *testing.T) {
	kv := NewKV()

	input := map[string]Value{
		"1": {ezt.IFile{"a", "/a", 1}, []string{"1", "2", "3"}},
		"2": {ezt.IFile{"b", "/b", 2}, []string{"2", "3"}},
		"3": {ezt.IFile{"c", "/c", 3}, []string{"3"}},
	}

	for k, v := range input {
		for i := range v.Peers {
			kv.Add(k, v.IFile, v.Peers[i])
		}
	}

	var wg sync.WaitGroup
	c := 1000
	wg.Add(c)
	for i := 0; i < c; i++ {
		go del(kv, &wg)
	}
	wg.Wait()

	t.Log("kv aprox. size:", kv.Stat())
}

func del(kv *KV, wg *sync.WaitGroup) {
	defer wg.Done()

	input := map[string][]string{
		"1": []string{"2", "3"},
		"2": []string{"3"},
		"3": []string{"3"},
	}

	expected := Values{
		{ezt.IFile{"a", "/a", 1}, []string{"1"}},
		{ezt.IFile{"b", "/b", 2}, []string{"2"}},
	}

	for k, v := range input {
		for i := range v {
			kv.Del(k, v[i])
		}
	}

	values := kv.List()

	result := Values(values)
	if !result.equals(expected) {
		log.Println("expected:", expected)
		log.Println("result:", result)
		log.Fatal("result != expected")
	}
}

func TestGet(t *testing.T) {
	kv := NewKV()

	input := map[string]Value{
		"1": {ezt.IFile{"a", "/a", 1}, []string{"1", "2", "3"}},
		"2": {ezt.IFile{"b", "/b", 2}, []string{"2", "3"}},
		"3": {ezt.IFile{"c", "/c", 3}, []string{"3"}},
	}

	for k, v := range input {
		for i := range v.Peers {
			kv.Add(k, v.IFile, v.Peers[i])
		}
	}

	var wg sync.WaitGroup
	c := 1000
	wg.Add(c)
	for i := 0; i < c; i++ {
		go get(kv, &wg)
	}
	wg.Wait()

	t.Log("kv aprox. size:", kv.Stat())
}

func get(kv *KV, wg *sync.WaitGroup) {
	defer wg.Done()

	input := []string{"1", "2", "3", "key not found"}
	expected := Values{
		{ezt.IFile{"a", "/a", 1}, []string{"1", "2", "3"}},
		{ezt.IFile{"b", "/b", 2}, []string{"2", "3"}},
		{ezt.IFile{"c", "/c", 3}, []string{"3"}},
	}
	var result Values
	for i := range input {
		v, err := kv.Get(input[i])
		if err != nil {
			continue
		}
		result = append(result, v)
	}

	if !result.equals(expected) {
		log.Println("expected:", expected)
		log.Println("result:", result)
		log.Fatal("result != expected")
	}
}

type Values []Value

func (l Values) equals(r Values) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		j := 0
		for ; j < len(r); j++ {
			lpeers := Peers(l[i].Peers)
			rpeers := Peers(r[j].Peers)
			if l[i].IFile.Equals(r[j].IFile) && lpeers.equals(rpeers) {
				break
			}
		}
		if j == len(r) {
			return false
		}
	}
	return true
}

type Peers []string

func (l Peers) equals(r Peers) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}
