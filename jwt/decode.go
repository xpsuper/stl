package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

func decode(token []byte) (header map[string]interface{}, payload map[string]interface{}, err error) {
	segments := bytes.Split(token, periodBytes)

	if len(segments) != 3 {
		return nil, nil, JwtErrInvalidToken
	}

	if header, err = decodeSegment(segments[0]); err != nil {
		return nil, nil, err
	}

	if payload, err = decodeSegment(segments[1]); err != nil {
		return nil, nil, err
	}

	return header, payload, nil
}

func decodeSegment(segment []byte) (m map[string]interface{}, err error) {
	s, err := base64.StdEncoding.DecodeString(string(segment))

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(s, &m); err != nil {
		return nil, err
	}

	return
}
