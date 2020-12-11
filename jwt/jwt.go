package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"time"
)

type XPJwtImpl struct {

}

type JwtSignOption struct {
	SignType   JwtAlgorithm   //签名算法
	Expiration time.Duration  //过期时间
	Audience   string         //接收方
	Issuer     string         //签发者
	Subject    string         //所面向的用户
	Header     JwtHeader      //自定义的头，将被合并至 Token 的头部
}

type JwtVerifyOption struct {
	SignType         JwtAlgorithm   //签名算法
	IngoreExpiration bool           //是否忽略到期时间
	Audience         string         //接收方
	Issuer           string         //签发方
	Subject          string         //所面向的用户
	Timeout          time.Duration  //检查到期时间时指定的时间容忍值
}

// 根据 payload 和 secret(私钥) 生成 JSON Web Token
// 当使用 HMAC 算法时，secret 为 string 或 []byte
// 当使用 RSA  算法时, secret 为 rsa.PrivateKey
// 如果 opt 为 nil，则默认使用 HS256 算法
func (jwt *XPJwtImpl) Sign(payload JwtPayload, secret interface{}, opt *JwtSignOption) (token []byte, err error) {
	if payload == nil {
		return nil, JwtErrEmptyPayload
	}

	if opt == nil {
		opt = &JwtSignOption{}
	}

	if secret == nil {
		return nil, JwtErrEmptySecretOrPrivateKey
	}

	var headerJSON, payloadJSON, signature []byte

	if headerJSON, err = marshalHeader(opt); err != nil {
		return
	}

	hBase64 := []byte(base64.StdEncoding.EncodeToString(headerJSON))

	if payloadJSON, err = marshalPayload(payload, opt); err != nil {
		return
	}

	pBase64 := []byte(base64.StdEncoding.EncodeToString(payloadJSON))

	if opt.SignType == "" {
		opt.SignType = JwtHS256
	}

	algImp, ok := algImpMap[opt.SignType]

	if !ok {
		return nil, JwtErrInvalidAlgorithm
	}

	if signature, err = algImp.sign(bytes.Join([][]byte{hBase64, pBase64},
		periodBytes), secret); err != nil {
		return
	}

	sigBase64 := []byte(base64.StdEncoding.EncodeToString(signature))

	return bytes.Join([][]byte{hBase64, pBase64, sigBase64}, periodBytes), nil
}

// 验证 token 并返回 header 和 payload
// 当使用 HMAC 算法时，secret 为 string 或 []byte
// 当使用 RSA  算法时, secret 为 rsa.PrivateKey
// 如果 opt 为 nil，则默认使用 HS256 算法
func (jwt *XPJwtImpl) Verify(token []byte, secret interface{}, opt *JwtVerifyOption) (header JwtHeader, payload JwtPayload, err error) {
	var (
		ok bool
		ai algorithmImplementation
	)

	if opt == nil {
		opt = &JwtVerifyOption{}
		opt.IngoreExpiration = true
	}

	if opt.SignType == "" {
		opt.SignType = JwtHS256
	}

	if ai, ok = algImpMap[opt.SignType]; !ok {
		return nil, nil, JwtErrInvalidAlgorithm
	}

	if header, payload, err = ai.verify(token, secret); err != nil {
		return nil, nil, JwtErrInvalidSignature
	}

	if !header.hasValidType() {
		return nil, nil, JwtErrInvalidHeaderType
	}

	if !payload.checkStringClaim("aud", opt.Audience) ||
		!payload.checkStringClaim("iss", opt.Issuer) ||
		!payload.checkStringClaim("sub", opt.Subject) {
		return nil, nil, JwtErrInvalidReservedClaim
	}

	if !opt.IngoreExpiration {
		if ok := payload.checkExpiration(opt.Timeout); !ok {
			return nil, nil, JwtErrTokenExpired
		}
	}

	return
}

func marshalHeader(opt *JwtSignOption) ([]byte, error) {
	h := map[string]interface{}{
		"alg": opt.SignType,
		"typ": "JWT",
	}

	if opt.Header != nil {
		if err := Map(&h, opt.Header); err != nil {
			return nil, err
		}
	}

	return json.Marshal(h)
}

func marshalPayload(payload JwtPayload, opt *JwtSignOption) ([]byte, error) {
	claims := JwtPayload{"iat": time.Now().Unix()}

	if opt.Issuer != "" {
		claims["iss"] = opt.Issuer
	}
	if opt.Expiration != 0 {
		claims["exp"] = opt.Expiration / 1e9
	}
	if opt.Subject != "" {
		claims["sub"] = opt.Subject
	}
	if opt.Audience != "" {
		claims["aud"] = opt.Audience
	}

	if err := Map(&claims, payload); err != nil {
		return nil, err
	}

	return json.Marshal(claims)
}