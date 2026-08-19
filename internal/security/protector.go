package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const encryptedPrefix = "forge:enc:v1:"

type Protector struct{ aead cipher.AEAD }

func NewProtector(configDir string) (*Protector, error) {
	keyPath := filepath.Join(configDir, "master.key")
	key, err := os.ReadFile(keyPath)
	if os.IsNotExist(err) {
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		if err = os.WriteFile(keyPath, key, 0600); err != nil {
			return nil, fmt.Errorf("criar chave local: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("ler chave local: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("chave local inválida")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Protector{aead: aead}, nil
}
func (p *Protector) Encrypt(value string) (string, error) {
	if value == "" || strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := p.aead.Seal(nonce, nonce, []byte(value), nil)
	return encryptedPrefix + base64.RawStdEncoding.EncodeToString(sealed), nil
}
func (p *Protector) Decrypt(value string) (string, error) {
	if !strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, encryptedPrefix))
	if err != nil {
		return "", err
	}
	n := p.aead.NonceSize()
	if len(raw) < n {
		return "", fmt.Errorf("valor secreto inválido")
	}
	plain, err := p.aead.Open(nil, raw[:n], raw[n:], nil)
	return string(plain), err
}
