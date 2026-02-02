package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go-proxy-manager/internal/model"
)

type Fetcher interface {
	Fetch(ctx context.Context, source model.Source) ([]*model.Proxy, error)
}

type CommonFetcher struct {
	client *http.Client
}

func NewCommonFetcher() *CommonFetcher {
	return &CommonFetcher{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

var protocolIPPortRegex = regexp.MustCompile(`(?:(http|https|socks5|socks4)://)?(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})[^\d]+(\d{1,5})`)

func (f *CommonFetcher) Fetch(ctx context.Context, source model.Source) ([]*model.Proxy, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, err
	}

	// Some providers require User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if source.Type == "json" {
		return f.parseJSON(body, source.URL)
	}

	return f.parseText(body, source.URL)
}

func (f *CommonFetcher) parseText(body []byte, sourceURL string) ([]*model.Proxy, error) {
	var proxies []*model.Proxy
	matches := protocolIPPortRegex.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		protocol := "http" // Default protocol
		if match[1] != "" {
			protocol = match[1]
		}
		port, _ := strconv.Atoi(match[3])
		if port > 0 && port <= 65535 {
			proxies = append(proxies, &model.Proxy{
				URL:      fmt.Sprintf("%s://%s:%s", protocol, match[2], match[3]),
				IP:       match[2],
				Port:     port,
				Protocol: protocol,
				Source:   sourceURL,
			})
		}
	}
	return proxies, nil
}

func (f *CommonFetcher) parseJSON(body []byte, sourceURL string) ([]*model.Proxy, error) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var proxies []*model.Proxy
	f.extractFromJSON(data, &proxies, sourceURL)
	return proxies, nil
}

func (f *CommonFetcher) extractFromJSON(data interface{}, proxies *[]*model.Proxy, sourceURL string) {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			f.extractFromJSON(item, proxies, sourceURL)
		}
	case map[string]interface{}:
		ip, okIP := v["ip"].(string)
		if !okIP {
			ip, okIP = v["ipAddress"].(string)
		}

		var port int
		if p, ok := v["port"].(float64); ok {
			port = int(p)
		} else if ps, ok := v["port"].(string); ok {
			port, _ = strconv.Atoi(ps)
		}

		protocol := "http" // Default protocol for JSON
		if p, ok := v["protocol"].(string); ok {
			protocol = p
		}

		if okIP && port > 0 {
			*proxies = append(*proxies, &model.Proxy{
				URL:      fmt.Sprintf("%s://%s:%d", protocol, ip, port),
				IP:       ip,
				Port:     port,
				Protocol: protocol,
				Source:   sourceURL,
			})
		} else {
			// Recursively search
			for _, val := range v {
				f.extractFromJSON(val, proxies, sourceURL)
			}
		}
	}
}
