package geobus

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	FilePath = "/etc/geolocation"
)

type GeolocationFileProvider struct {
	name   string
	result Result
	path   string
	period time.Duration
	ttl    time.Duration
}

func NewGeolocationFileProvider() *GeolocationFileProvider {
	return &GeolocationFileProvider{
		name:   "GeolocationFile",
		path:   FilePath,
		period: 10 * time.Second,
		ttl:    2 * time.Minute,
	}
}

func (p *GeolocationFileProvider) Name() string {
	return p.name
}

func (p *GeolocationFileProvider) LookupStream(ctx context.Context, key string) <-chan Result {
	out := make(chan Result)
	go func() {
		defer close(out)
		var (
			lastLat, lastLon float64
			lastAlt, lastAcc float64
			haveLast         bool
		)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			lat, lon, alt, acc, err := p.readFile()
			if err != nil {
				// File missing or malformed — just retry later
				time.Sleep(p.period)
				continue
			}

			// Only emit if values changed or it's the first read
			if !haveLast || lat != lastLat || lon != lastLon || alt != lastAlt || acc != lastAcc {
				lastLat, lastLon, lastAlt, lastAcc = lat, lon, alt, acc
				haveLast = true

				r := Result{
					Key:            key,
					Lat:            lat,
					Lon:            lon,
					AccuracyMeters: acc,
					Confidence:     1.0,
					Source:         p.name,
					At:             time.Now(),
					TTL:            p.ttl,
				}

				select {
				case <-ctx.Done():
					return
				case out <- r:
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(p.period):
			}
		}
	}()
	return out
}

// readFile parses the four-line format (lat, lon, alt, acc) from the configured file.
func (p *GeolocationFileProvider) readFile() (lat, lon, alt, acc float64, err error) {
	f, err := os.Open(p.path)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("open %s: %w", p.path, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close geolocation file: %w", closeErr))
		}
	}()

	scanner := bufio.NewScanner(f)
	var values []float64
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		v, parseErr := strconv.ParseFloat(line, 64)
		if parseErr != nil {
			return 0, 0, 0, 0, fmt.Errorf("invalid number in %s: %w", p.path, parseErr)
		}
		values = append(values, v)
	}
	if len(values) < 4 {
		return 0, 0, 0, 0, errors.New("geolocation file missing required lines (need 4)")
	}
	return values[0], values[1], values[2], values[3], nil
}
