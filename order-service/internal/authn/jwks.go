package authn

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type keySet struct {
	client    *http.Client
	url       string
	ttl       time.Duration
	mu        sync.RWMutex
	expiresAt time.Time
	keys      map[string]*rsa.PublicKey
}

func newKeySet(url string) *keySet {
	return &keySet{
		client: &http.Client{Timeout: 5 * time.Second},
		url:    url,
		ttl:    5 * time.Minute,
		keys:   map[string]*rsa.PublicKey{},
	}
}

func (ks *keySet) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if key := ks.cached(kid); key != nil {
		return key, nil
	}

	if err := ks.refresh(ctx); err != nil {
		return nil, err
	}

	if key := ks.cached(kid); key != nil {
		return key, nil
	}
	return nil, fmt.Errorf("jwks key %q not found", kid)
}

func (ks *keySet) cached(kid string) *rsa.PublicKey {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if now().After(ks.expiresAt) {
		return nil
	}
	return ks.keys[kid]
}

func (ks *keySet) refresh(ctx context.Context) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if now().Before(ks.expiresAt) && len(ks.keys) > 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ks.url, nil)
	if err != nil {
		return err
	}

	resp, err := ks.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("jwks request failed with status %d", resp.StatusCode)
	}

	var body jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(body.Keys))
	for _, jwk := range body.Keys {
		key, err := jwk.rsaPublicKey()
		if err != nil {
			continue
		}
		keys[jwk.Kid] = key
	}
	if len(keys) == 0 {
		return errors.New("jwks did not contain usable rsa keys")
	}

	ks.keys = keys
	ks.expiresAt = now().Add(ks.ttl)
	return nil
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (j jwkKey) rsaPublicKey() (*rsa.PublicKey, error) {
	if j.Kid == "" || j.Kty != "RSA" || j.N == "" || j.E == "" {
		return nil, errors.New("unsupported jwk")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, err
	}

	exponent := 0
	for _, b := range eBytes {
		exponent = exponent<<8 + int(b)
	}
	if exponent == 0 {
		return nil, errors.New("invalid rsa exponent")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: exponent,
	}, nil
}
