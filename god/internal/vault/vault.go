package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Vault struct {
	mu       sync.RWMutex
	dataDir  string
	path     string
	keyPath  string
	secrets  map[string]string
}

func New(dataDir string) *Vault {
	return &Vault{
		dataDir: dataDir,
		path:    filepath.Join(dataDir, "credentials.vault"),
		keyPath: filepath.Join(dataDir, ".vault-key"),
		secrets: make(map[string]string),
	}
}

func (v *Vault) Load() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	key, err := v.loadOrCreateKey()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(v.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	decrypted, err := decrypt(key, data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decrypted, &v.secrets)
}

func (v *Vault) Set(provider, secret string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.secrets[provider] = secret
	return v.save()
}

func (v *Vault) Get(provider string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	s, ok := v.secrets[provider]
	return s, ok
}

func (v *Vault) Delete(provider string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.secrets, provider)
	return v.save()
}

func (v *Vault) Connected() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	out := make([]string, 0, len(v.secrets))
	for k := range v.secrets {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (v *Vault) Masked(provider string) string {
	s, ok := v.Get(provider)
	if !ok || s == "" {
		return ""
	}
	if len(s) < 8 {
		return "***"
	}
	return s[:3] + "..." + s[len(s)-4:]
}

func (v *Vault) save() error {
	data, err := json.Marshal(v.secrets)
	if err != nil {
		return err
	}

	key, err := v.loadOrCreateKey()
	if err != nil {
		return err
	}

	encrypted, err := encrypt(key, data)
	if err != nil {
		return err
	}

	return os.WriteFile(v.path, encrypted, 0600)
}

func (v *Vault) loadOrCreateKey() ([]byte, error) {
	if data, err := os.ReadFile(v.keyPath); err == nil && len(data) == 32 {
		return data, nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(v.keyPath, key, 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	mac := hmacSHA256(key, plaintext)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return append(mac, append(nonce, ciphertext...)...), nil
}

func decrypt(key, data []byte) ([]byte, error) {
	if len(data) < 32+12+16 {
		return nil, errors.New("invalid vault data")
	}
	storedMAC := data[:32]
	nonce := data[32 : 32+12]
	ciphertext := data[32+12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(storedMAC, hmacSHA256(key, plaintext)) {
		return nil, errors.New("vault integrity check failed")
	}
	return plaintext, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
