package authn

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Verifier struct {
	cfg  Config
	keys *keySet
}

func NewVerifier(cfg Config) (*Verifier, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return &Verifier{cfg: cfg}, nil
	}
	return &Verifier{
		cfg:  cfg,
		keys: newKeySet(cfg.JWKSURL),
	}, nil
}

func (v *Verifier) Verify(ctx context.Context, token string) (Claims, error) {
	if !v.cfg.Enabled {
		return Claims{Subject: "dev", Role: RoleAdmin, Roles: []string{string(RoleAdmin)}}, nil
	}

	header, payload, signingInput, signature, err := splitToken(token)
	if err != nil {
		return Claims{}, err
	}

	if alg, _ := header["alg"].(string); alg != "RS256" {
		return Claims{}, errors.New("unsupported token algorithm")
	}
	kid, _ := header["kid"].(string)
	if kid == "" {
		return Claims{}, errors.New("token is missing kid")
	}

	key, err := v.keys.key(ctx, kid)
	if err != nil {
		return Claims{}, err
	}
	if err := verifyRS256(key, signingInput, signature); err != nil {
		return Claims{}, err
	}
	if err := v.validatePayload(payload); err != nil {
		return Claims{}, err
	}

	claims := claimsFromPayload(payload, v.cfg)
	if claims.Subject == "" {
		return Claims{}, errors.New("token subject is missing")
	}
	return claims, nil
}

func splitToken(token string) (map[string]any, map[string]any, string, []byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, "", nil, errors.New("malformed bearer token")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, "", nil, err
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, "", nil, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, nil, "", nil, err
	}

	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, "", nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil, "", nil, err
	}

	return header, payload, parts[0] + "." + parts[1], signature, nil
}

func verifyRS256(key *rsa.PublicKey, signingInput string, signature []byte) error {
	hash := sha256.Sum256([]byte(signingInput))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature)
}

func (v *Verifier) validatePayload(payload map[string]any) error {
	if issuer := strings.TrimRight(stringClaim(payload, "iss"), "/"); issuer != v.cfg.Issuer {
		return fmt.Errorf("invalid token issuer")
	}
	if !audienceMatches(payload["aud"], v.cfg.Audience) {
		return fmt.Errorf("invalid token audience")
	}
	if exp, ok := numericDate(payload["exp"]); !ok || now().After(exp) {
		return errors.New("token is expired")
	}
	if nbf, ok := numericDate(payload["nbf"]); ok && now().Before(nbf) {
		return errors.New("token is not active yet")
	}
	return nil
}

func audienceMatches(raw any, expected string) bool {
	switch value := raw.(type) {
	case string:
		return value == expected
	case []any:
		for _, item := range value {
			if audience, ok := item.(string); ok && audience == expected {
				return true
			}
		}
	}
	return false
}

func numericDate(raw any) (time.Time, bool) {
	value, ok := raw.(float64)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(int64(value), 0), true
}
