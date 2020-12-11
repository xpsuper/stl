package jwt

import (
	"errors"
	"time"
)

// Algorithm represents a supported hash algorithms.
type JwtAlgorithm string

const (
	// HS256 represents HMAC using SHA-256 hash algorithm.
	JwtHS256 JwtAlgorithm = "HS256"
	// HS384 represents HMAC using SHA-384 hash algorithm.
	JwtHS384 JwtAlgorithm = "HS384"
	// HS512 represents HMAC using SHA-512 hash algorithm.
	JwtHS512 JwtAlgorithm = "HS512"
	// RS256 represents RSASSA using SHA-256 hash algorithm.
	JwtRS256 JwtAlgorithm = "RS256"
	// RS384 represents RSASSA using SHA-384 hash algorithm.
	JwtRS384 JwtAlgorithm = "RS384"
	// RS512 represents RSASSA using SHA-512 hash algorithm.
	JwtRS512 JwtAlgorithm = "RS512"
)

var (
	// ErrEmptyPayload is returned when the payload given to Sign is empty.
	JwtErrEmptyPayload = errors.New("jwt: empty payload")
	// ErrEmptySecretOrPrivateKey is returned when the secret or private key
	// given is empy.
	JwtErrEmptySecretOrPrivateKey = errors.New("jwt: empty secret or private key")
	// ErrInvalidKeyType is returned when the type of given key is wrong.
	JwtErrInvalidKeyType = errors.New("jwt: invalid key")
	// ErrInvalidSignature is returned when the given signature is invalid.
	JwtErrInvalidSignature = errors.New("jwt: invalid signature")
	// ErrInvalidHeaderType is returned when "typ" not found in header and is not
	// "JWT".
	JwtErrInvalidHeaderType = errors.New("jwt: invalid header type")
	// ErrInvalidToken is returned when the formation of the token is not
	// "XXX.XXX.XXX".
	JwtErrInvalidToken = errors.New("jwt: invalid token")
	// ErrInvalidAlgorithm is returned when the algorithm is not support.
	JwtErrInvalidAlgorithm = errors.New("jwt: invalid algorithm")
	// ErrInvalidReservedClaim is returned when the reserved claim dose not match
	// with the given value in VerifyOption.
	JwtErrInvalidReservedClaim = errors.New("jwt: invalid reserved claim")
	// ErrPayloadMissingIat is returned when the payload is missing "iat".
	JwtErrPayloadMissingIat = errors.New("jwt: payload missing iat")
	// ErrPayloadMissingExp is returned when the payload is missing "exp".
	JwtErrPayloadMissingExp = errors.New("jwt: payload missing exp")
	// ErrTokenExpired is returned when the token is expired.
	JwtErrTokenExpired = errors.New("jwt: token expired")

	periodBytes = []byte(".")
	algImpMap   = map[JwtAlgorithm]algorithmImplementation{}
)

type algorithmImplementation interface {
	sign(content []byte, key interface{}) ([]byte, error)
	verify(signing []byte, key interface{}) (JwtHeader, JwtPayload, error)
}

// Header represents a JWT header.
type JwtHeader map[string]interface{}

func (h JwtHeader) hasValidType() bool {
	var (
		typ      interface{}
		received string
		ok       bool
	)

	if typ, ok = h["typ"]; !ok {
		return false
	}

	if received, ok = typ.(string); !ok {
		return false
	}

	return received == "JWT"
}

type JwtPayload map[string]interface{}

func (p JwtPayload) checkStringClaim(key, expected string) bool {
	if expected == "" {
		return true
	}

	var (
		received string
		v        interface{}
		ok       bool
	)

	if v, ok = p[key]; !ok {
		return false
	}

	if received, ok = v.(string); !ok {
		return false
	}

	return expected == received
}

func (p JwtPayload) iat() (t time.Time, err error) {
	var (
		iat float64
		ok  bool
		v   interface{}
	)

	if v, ok = p["iat"]; !ok {
		return t, JwtErrPayloadMissingIat
	}

	if iat, ok = v.(float64); !ok {
		return t, JwtErrPayloadMissingIat
	}

	return time.Unix(int64(iat), 0), nil
}

func (p JwtPayload) expTime() (t time.Time, err error) {
	var (
		exp float64
		iat time.Time
		ok  bool
		v   interface{}
	)

	if v, ok = p["exp"]; !ok {
		return t, JwtErrPayloadMissingExp
	}

	if exp, ok = v.(float64); !ok {
		return t, JwtErrPayloadMissingExp
	}

	if iat, err = p.iat(); err != nil {
		return
	}

	return iat.Add(time.Duration(int64(exp * 1e9))), nil
}

func (p JwtPayload) checkExpiration(timeout time.Duration) bool {
	if exp, err := p.expTime(); err == nil {
		return time.Now().Add(timeout).Before(exp)
	}

	return false
}
