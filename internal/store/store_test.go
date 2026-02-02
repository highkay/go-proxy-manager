package store

import (
	"testing"
	"time"

	"go-proxy-manager/internal/model"
)

func TestStore_GetSorted(t *testing.T) {
	s := NewStore()

	p1 := &model.Proxy{URL: "http://1.1.1.1:80", Latency: 200 * time.Millisecond}
	p2 := &model.Proxy{URL: "http://2.2.2.2:80", Latency: 100 * time.Millisecond}
	p3 := &model.Proxy{URL: "http://3.3.3.3:80", Latency: 300 * time.Millisecond}

	s.Add(p1)
	s.Add(p2)
	s.Add(p3)

	sorted := s.GetSorted(0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 proxies, got %d", len(sorted))
	}

	if sorted[0].URL != "http://2.2.2.2:80" {
		t.Errorf("expected fastest proxy first, got %s", sorted[0].URL)
	}

	if sorted[2].URL != "http://3.3.3.3:80" {
		t.Errorf("expected slowest proxy last, got %s", sorted[2].URL)
	}
}

func TestStore_Limit(t *testing.T) {
	s := NewStore()

	s.Add(&model.Proxy{URL: "http://1.1.1.1:80", Latency: 100 * time.Millisecond})
	s.Add(&model.Proxy{URL: "http://2.2.2.2:80", Latency: 200 * time.Millisecond})

	limited := s.GetSorted(1)
	if len(limited) != 1 {
		t.Errorf("expected 1 proxy, got %d", len(limited))
	}
}
