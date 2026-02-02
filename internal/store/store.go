package store

import (
	"sort"
	"sync"

	"go-proxy-manager/internal/model"
)

type Store struct {
	mu      sync.RWMutex
	proxies map[string]*model.Proxy
}

func NewStore() *Store {
	return &Store{
		proxies: make(map[string]*model.Proxy),
	}
}

func (s *Store) Add(p *model.Proxy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxies[p.URL] = p
}

func (s *Store) Remove(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.proxies, url)
}

func (s *Store) GetSorted(limit int) []*model.Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*model.Proxy, 0, len(s.proxies))
	for _, p := range s.proxies {
		list = append(list, p)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Latency < list[j].Latency
	})

	if limit > 0 && limit < len(list) {
		return list[:limit]
	}
	return list
}

func (s *Store) GetAll() []*model.Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*model.Proxy, 0, len(s.proxies))
	for _, p := range s.proxies {
		list = append(list, p)
	}
	return list
}
