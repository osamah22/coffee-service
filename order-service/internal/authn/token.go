package authn

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type TokenIssuer struct {
	cfg Config
}

func NewTokenIssuer(cfg Config) (*TokenIssuer, error) {
	if !cfg.Enabled {
		return &TokenIssuer{cfg: cfg}, nil
	}
	if strings.TrimSpace(cfg.TokenSecret) == "" {
		return nil, errors.New("auth jwt secret is required")
	}
	return &TokenIssuer{cfg: cfg}, nil
}

func (i *TokenIssuer) Issue(subject, email string, role Role) (string, error) {
	now := now()
	payload := map[string]any{
		"iss":    i.cfg.Issuer,
		"aud":    i.cfg.Audience,
		"sub":    subject,
		"email":  email,
		"groups": []string{string(role)},
		"iat":    now.Unix(),
		"nbf":    now.Unix(),
		"exp":    now.Add(24 * time.Hour).Unix(),
	}
	return signHS256(i.cfg.TokenSecret, payload)
}

func signHS256(secret string, payload map[string]any) (string, error) {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerPart, err := encodeJSONPart(header)
	if err != nil {
		return "", err
	}
	payloadPart, err := encodeJSONPart(payload)
	if err != nil {
		return "", err
	}
	signingInput := headerPart + "." + payloadPart
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + signature, nil
}

func encodeJSONPart(value any) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
