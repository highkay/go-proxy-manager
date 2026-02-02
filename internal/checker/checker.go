package checker

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go-proxy-manager/internal/model"

	"golang.org/x/net/proxy"
)

type Checker struct {
	targets []string
	timeout time.Duration
	workers int
}

func NewChecker(targets []string, timeout time.Duration, workers int) *Checker {
	return &Checker{
		targets: targets,
		timeout: timeout,
		workers: workers,
	}
}

func (c *Checker) Check(ctx context.Context, p *model.Proxy) bool {
	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		return false
	}

	var transport *http.Transport

	if proxyURL.Scheme == "socks5" {
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
		if err != nil {
			return false
		}
		transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   c.timeout,
	}

	// Try target URLs until one works
	for _, target := range c.targets {
		start := time.Now()
		req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
		if err != nil {
			continue
		}

		// Disguise as a browser
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			p.Latency = time.Since(start)
			p.LastCheck = time.Now()
			p.FailCount = 0
			return true
		}
	}

	p.FailCount++
	return false
}

// WorkerPool processes proxies from input channel and sends valid ones to output channel
func (c *Checker) RunWorkerPool(ctx context.Context, input <-chan *model.Proxy, output chan<- *model.Proxy) {
	var wg sync.WaitGroup
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-input:
					if !ok {
						return
					}
					if c.Check(ctx, p) {
						output <- p
					}
				}
			}
		}()
	}
	wg.Wait()
	close(output)
}
