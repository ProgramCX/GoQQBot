package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(body string) *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestTokenTypesConcurrentCacheAndGateway(t *testing.T) {
	for _, expiry := range []string{`7200`, `"7200"`} {
		t.Run(expiry, func(t *testing.T) {
			var calls atomic.Int32
			client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "api.bot.qq.com" {
					t.Errorf("wrong host: %s", r.URL.Host)
				}
				if r.URL.Path == "/app/getAppAccessToken" {
					calls.Add(1)
					return response(`{"access_token":"test-token","expires_in":` + expiry + `}`), nil
				}
				if r.URL.Path != "/gateway" || r.Header.Get("Authorization") != "QQBot test-token" {
					t.Error("bad gateway request")
				}
				return response(`{"url":"wss://example.test/gateway"}`), nil
			})}))
			var wg sync.WaitGroup
			for i := 0; i < 12; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if _, err := client.AccessToken(context.Background()); err != nil {
						t.Error(err)
					}
				}()
			}
			wg.Wait()
			if calls.Load() != 1 {
				t.Fatalf("token requests: %d", calls.Load())
			}
			if endpoint, err := client.Gateway(context.Background()); err != nil || endpoint != "wss://example.test/gateway" {
				t.Fatalf("Gateway: %s %v", endpoint, err)
			}
		})
	}
}
func TestTokenBusinessErrorAndCancellation(t *testing.T) {
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return response(`{"code":100016,"message":"invalid credentials"}`), nil
	})}))
	_, err := client.AccessToken(context.Background())
	var failure *APIError
	if !errors.As(err, &failure) || failure.Code != 100016 {
		t.Fatalf("token failure: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.AccessToken(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestAPIErrorPreservesTrace(t *testing.T) {
	client := NewClient("test", "secret", WithHTTPClient(&http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/app/getAppAccessToken" {
			return response(`{"access_token":"token","expires_in":7200}`), nil
		}
		return response(`{"err_code":40034025,"trace_id":"trace"}`), nil
	})}))
	_, err := client.Get(context.Background(), "/gateway")
	var failure *APIError
	if !errors.As(err, &failure) || failure.Code != 40034025 || failure.TraceID != "trace" {
		t.Fatalf("failure: %v", err)
	}
}
