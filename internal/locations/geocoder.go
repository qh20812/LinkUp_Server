package locations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	nominatimEndpoint = "https://nominatim.openstreetmap.org/reverse"
	nominatimUA       = "LinkUp/1.0 (Vietnamese social network; profile location feature)"
	minRequestGap     = 1100 * time.Millisecond
	httpTimeout       = 5 * time.Second
)

// Geocoder resolves geographic coordinates into a human-readable place name.
type Geocoder interface {
	ReverseDisplayName(ctx context.Context, lat, lon float64) (string, error)
}

// NominatimGeocoder is a polite Nominatim client: it enforces the OSM usage
// policy (≤1 request/second, descriptive User-Agent, HTTP timeout) and caches
// results by rounded coordinates.
type NominatimGeocoder struct {
	client      *http.Client
	mu          sync.Mutex
	lastRequest time.Time
	cache       map[string]*string
}

func NewNominatimGeocoder(client *http.Client) *NominatimGeocoder {
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}
	return &NominatimGeocoder{
		client:      client,
		cache:       make(map[string]*string),
	}
}

// ReverseDisplayName returns the display_name for the given coordinates.
// An empty string means no usable address was found.
func (g *NominatimGeocoder) ReverseDisplayName(ctx context.Context, lat, lon float64) (string, error) {
	key := cacheKey(lat, lon)

	g.mu.Lock()
	if v, ok := g.cache[key]; ok {
		g.mu.Unlock()
		if v == nil {
			return "", nil
		}
		return *v, nil
	}
	// Throttle — at most one request per minRequestGap.
	if wait := minRequestGap - time.Since(g.lastRequest); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			g.mu.Unlock()
			return "", ctx.Err()
		}
	}
	g.lastRequest = time.Now()
	g.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nominatimEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build reverse request: %w", err)
	}
	q := req.URL.Query()
	q.Set("format", "jsonv2")
	q.Set("lat", fmt.Sprintf("%.6f", lat))
	q.Set("lon", fmt.Sprintf("%.6f", lon))
	q.Set("zoom", "16")
	q.Set("accept-language", "vi")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", nominatimUA)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("nominatim reverse: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read nominatim response: %w", err)
	}
	// Cache network failures so we do not hammer OSM on repeated clicks.
	g.mu.Lock()
	defer g.mu.Unlock()
	if resp.StatusCode != http.StatusOK {
		g.cache[key] = nil
		return "", fmt.Errorf("nominatim http %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var payload struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		g.cache[key] = nil
		return "", fmt.Errorf("parse nominatim response: %w", err)
	}
	if payload.DisplayName == "" {
		g.cache[key] = nil
		return "", nil
	}
	copy := payload.DisplayName
	g.cache[key] = &copy
	return payload.DisplayName, nil
}

func cacheKey(lat, lon float64) string {
	return fmt.Sprintf("%.4f,%.4f", lat, lon)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}