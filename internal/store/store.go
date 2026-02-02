package store

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func (s *Store) Save(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*model.Proxy, 0, len(s.proxies))
	for _, p := range s.proxies {
		list = append(list, p)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(list)
}

func (s *Store) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var list []*model.Proxy
	if err := json.NewDecoder(file).Decode(&list); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range list {
		s.proxies[p.URL] = p
	}
	return nil
}
