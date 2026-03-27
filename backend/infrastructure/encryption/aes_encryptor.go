package encryption

import (
  "crypto/aes"
  "crypto/cipher"
  "crypto/rand"
  "encoding/base64"
  "fmt"
  "io"

  tokenencryptor "github.com/Asheze1127/progress-checker/backend/application/service/token_encryptor"
)

// Compile-time check that AESEncryptor implements tokenencryptor.TokenEncryptor.
var _ tokenencryptor.TokenEncryptor = (*AESEncryptor)(nil)

// AESEncryptor encrypts and decrypts tokens using AES-GCM.
type AESEncryptor struct {
  key []byte
}

// NewAESEncryptor creates a new AESEncryptor with the given key.
// The key must be 16, 24, or 32 bytes long (for AES-128, AES-192, AES-256).
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
  keyLen := len(key)
  if keyLen != 16 && keyLen != 24 && keyLen != 32 {
    return nil, fmt.Errorf("encryption key must be 16, 24, or 32 bytes, got %d", keyLen)
  }

  return &AESEncryptor{key: key}, nil
}

// Encrypt encrypts the plaintext using AES-GCM and returns a base64-encoded ciphertext.
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
  block, err := aes.NewCipher(e.key)
  if err != nil {
    return "", fmt.Errorf("create cipher: %w", err)
  }

  aesGCM, err := cipher.NewGCM(block)
  if err != nil {
    return "", fmt.Errorf("create gcm: %w", err)
  }

  nonce := make([]byte, aesGCM.NonceSize())
  if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    return "", fmt.Errorf("generate nonce: %w", err)
  }

  ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

  return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext using AES-GCM.
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
  data, err := base64.StdEncoding.DecodeString(ciphertext)
  if err != nil {
    return "", fmt.Errorf("decode base64: %w", err)
  }

  block, err := aes.NewCipher(e.key)
  if err != nil {
    return "", fmt.Errorf("create cipher: %w", err)
  }

  aesGCM, err := cipher.NewGCM(block)
  if err != nil {
    return "", fmt.Errorf("create gcm: %w", err)
  }

  nonceSize := aesGCM.NonceSize()
  if len(data) < nonceSize {
    return "", fmt.Errorf("ciphertext too short")
  }

  nonce, encryptedData := data[:nonceSize], data[nonceSize:]
  plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
  if err != nil {
    return "", fmt.Errorf("decrypt: %w", err)
  }

  return string(plaintext), nil
}
