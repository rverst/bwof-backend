package helper

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenString(len int) (string, error) {
	if len > 128 {
		len = 128
	}
	b := make([]byte, len*2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	enc := base64.NewEncoder(base64.RawStdEncoding, buf)
	_, err = enc.Write(b)
	if err != nil {
		return "", err
	}
	err = enc.Close()
	if err != nil {
		return "", err
	}

	s1 := make([]byte, len)
	c := 0
	for _, b := range buf.Bytes() {
		if b == '+' || b == '/' {
			continue
		}
		s1[c] = b
		c++
		if c == len {
			break
		}
	}
	if c != len {
		return GenString(len)
	}
	return string(s1), nil
}

func genKey(len int) []byte {
	if len == 0 {
		return nil
	}
	c := len
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return b
}

func xorString(key, input string) (string, error) {
	k, err := hex.DecodeString(key)
	if err != nil {
		return "", err
	}
	i, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	r, err := xor(k, i)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(r), nil
}

func xor(key, input []byte) ([]byte, error) {
	if len(key) < len(input) {
		return nil, fmt.Errorf("len(key) must be >= len(input)")
	}

	out := make([]byte, len(input))
	for i, c := range input {
		out[i] = key[i] ^ c
	}

	return out, nil
}
