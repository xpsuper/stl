package jwt

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"hash"
)

func init() {
	algImpMap[JwtHS256] = hmacAlgImp{hashFunc: crypto.SHA256.New}
	algImpMap[JwtHS384] = hmacAlgImp{hashFunc: crypto.SHA384.New}
	algImpMap[JwtHS512] = hmacAlgImp{hashFunc: crypto.SHA512.New}
}

type hmacAlgImp struct {
	hashFunc func() hash.Hash
}

func (ha hmacAlgImp) sign(content []byte, secret interface{}) ([]byte, error) {
	var s []byte

	switch secret.(type) {
	case []byte:
		s = secret.([]byte)
	case string:
		s = []byte(secret.(string))
	default:
		return nil, JwtErrInvalidKeyType
	}

	h := hmac.New(ha.hashFunc, s)

	h.Write(content)

	return h.Sum(nil), nil
}

func (ha hmacAlgImp) verify(token []byte, secret interface{}) (header JwtHeader, payload JwtPayload, err error) {
	if header, payload, err = decode(token); err != nil {
		return
	}

	signatureReceive, err := base64.StdEncoding.DecodeString(string(bytes.Split(token, periodBytes)[2]))

	if err != nil {
		return
	}

	signatureExpect, err := ha.sign(token[0:bytes.LastIndexByte(token, '.')], secret)

	if err != nil {
		return
	}

	ok := hmac.Equal(signatureExpect, signatureReceive)

	if !ok {
		return nil, nil, JwtErrInvalidSignature
	}

	return
}