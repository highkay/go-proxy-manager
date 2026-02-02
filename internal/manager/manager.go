package manager

import (
	"context"
	"log/slog"
	"time"

	"go-proxy-manager/internal/checker"
	"go-proxy-manager/internal/config"
	"go-proxy-manager/internal/fetcher"
	"go-proxy-manager/internal/model"
	"go-proxy-manager/internal/store"
)

type Manager struct {
	cfg     *config.Config
	store   *store.Store
	fetcher fetcher.Fetcher
	checker *checker.Checker
}

func NewManager(cfg *config.Config, s *store.Store) *Manager {
	return &Manager{
		cfg:     cfg,
		store:   s,
		fetcher: fetcher.NewCommonFetcher(),
		checker: checker.NewChecker(cfg.Validation.TargetURLs, cfg.Validation.Timeout, cfg.App.ThreadCount),
	}
}

func (m *Manager) Start(ctx context.Context) {
	// 1. Start fetching for each source
	for _, src := range m.cfg.Sources {
		go m.fetchLoop(ctx, src)
	}

	// 2. Start periodic validation of existing proxies
	go m.checkLoop(ctx)
}

func (m *Manager) fetchLoop(ctx context.Context, source model.Source) {
	ticker := time.NewTicker(source.Interval)
	defer ticker.Stop()

	// Initial fetch
	m.runFetch(ctx, source)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runFetch(ctx, source)
		}
	}
}

func (m *Manager) runFetch(ctx context.Context, source model.Source) {
	var proxies []*model.Proxy
	var err error

	maxRetries := 3
	baseDelay := 2 * time.Second

	for i := 0; i <= maxRetries; i++ {
		slog.Info("fetching proxies", "source", source.URL, "attempt", i+1)
		proxies, err = m.fetcher.Fetch(ctx, source)
		if err == nil {
			break
		}

		slog.Warn("fetch failed", "source", source.URL, "attempt", i+1, "error", err)
		if i < maxRetries {
			delay := baseDelay * time.Duration(1<<i) // 2s, 4s, 8s
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		}
	}

	if err != nil {
		slog.Error("fetch permanently failed after retries", "source", source.URL, "error", err)
		return
	}

	// Filter duplicates and recently checked proxies
	uniqueProxies := make([]*model.Proxy, 0, len(proxies))
	seen := make(map[string]bool)

	// Snapshot of existing proxies to check for redundancy
	// Note: Store is thread-safe, but we iterate a copy or rely on map lookup.
	// Since Store doesn't expose a fast lookup without lock, we'll use GetAll for now
	// or improved Store API later. For now, we trust the fetching process is not too frequent.
	// Actually, let's just use GetAll() which is read-locked.
	existingProxies := m.store.GetAll()
	existingMap := make(map[string]*model.Proxy, len(existingProxies))
	for _, p := range existingProxies {
		existingMap[p.URL] = p
	}

	for _, p := range proxies {
		// 1. Deduplicate within this batch
		if seen[p.URL] {
			continue
		}
		seen[p.URL] = true

		// 2. Check if exists in store and recently checked
		if existing, ok := existingMap[p.URL]; ok {
			// If checked within last 10 minutes, skip
			if time.Since(existing.LastCheck) < 10*time.Minute {
				continue
			}
			// If fail count is high, maybe skip too? Let's keep trying for now but could optimize.
		}

		uniqueProxies = append(uniqueProxies, p)
	}

	slog.Info("fetched proxies", "raw_count", len(proxies), "unique_new_count", len(uniqueProxies), "source", source.URL)

	if len(uniqueProxies) == 0 {
		return
	}

	// Send to checker
	input := make(chan *model.Proxy, len(uniqueProxies))
	output := make(chan *model.Proxy, len(uniqueProxies))

	go func() {
		for _, p := range uniqueProxies {
			input <- p
		}
		close(input)
	}()

	go m.checker.RunWorkerPool(ctx, input, output)

	// Collect valid proxies
	go func() {
		for p := range output {
			m.store.Add(p)
			slog.Debug("added valid proxy", "url", p.URL, "latency", p.Latency)
		}
	}()
}

func (m *Manager) checkLoop(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.Validation.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.revalidateAll(ctx)
		}
	}
}

func (m *Manager) revalidateAll(ctx context.Context) {
	proxies := m.store.GetAll()
	if len(proxies) == 0 {
		return
	}

	slog.Info("revalidating all proxies", "count", len(proxies))

	input := make(chan *model.Proxy, len(proxies))
	output := make(chan *model.Proxy, len(proxies))

	go func() {
		for _, p := range proxies {
			input <- p
		}
		close(input)
	}()

	go m.checker.RunWorkerPool(ctx, input, output)

	// We don't need to do much here because Checker updates the Proxy object in-place
	// and Store holds the pointer. However, we might want to remove failed ones.
	// For simplicity, we'll just let them stay or be removed if FailCount is too high.
	go func() {
		for range output {
			// Proxy is updated by checker
		}
		// Check for expired or too many failures
		for _, p := range m.store.GetAll() {
			if p.FailCount > 3 {
				m.store.Remove(p.URL)
				slog.Info("removed dead proxy", "url", p.URL)
			}
		}
	}()
}
