package encryption_test

import (
  "testing"

  "github.com/Asheze1127/progress-checker/backend/infrastructure/encryption"
)

func TestAESEncryptor_RoundTrip(t *testing.T) {
  t.Parallel()

  key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
  encryptor, err := encryption.NewAESEncryptor(key)
  if err != nil {
    t.Fatalf("NewAESEncryptor() error = %v", err)
  }

  plaintext := "ghp_testtoken12345"

  ciphertext, err := encryptor.Encrypt(plaintext)
  if err != nil {
    t.Fatalf("Encrypt() error = %v", err)
  }

  if ciphertext == plaintext {
    t.Error("Encrypt() ciphertext should differ from plaintext")
  }

  decrypted, err := encryptor.Decrypt(ciphertext)
  if err != nil {
    t.Fatalf("Decrypt() error = %v", err)
  }

  if decrypted != plaintext {
    t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
  }
}

func TestAESEncryptor_DifferentCiphertexts(t *testing.T) {
  t.Parallel()

  key := []byte("0123456789abcdef0123456789abcdef")
  encryptor, err := encryption.NewAESEncryptor(key)
  if err != nil {
    t.Fatalf("NewAESEncryptor() error = %v", err)
  }

  plaintext := "ghp_testtoken12345"

  ct1, err := encryptor.Encrypt(plaintext)
  if err != nil {
    t.Fatalf("Encrypt() error = %v", err)
  }

  ct2, err := encryptor.Encrypt(plaintext)
  if err != nil {
    t.Fatalf("Encrypt() error = %v", err)
  }

  if ct1 == ct2 {
    t.Error("Encrypt() should produce different ciphertexts for same plaintext due to random nonce")
  }
}

func TestNewAESEncryptor_InvalidKeyLength(t *testing.T) {
  t.Parallel()

  _, err := encryption.NewAESEncryptor([]byte("short"))
  if err == nil {
    t.Error("NewAESEncryptor() should return error for invalid key length")
  }
}

func TestAESEncryptor_DecryptInvalidData(t *testing.T) {
  t.Parallel()

  key := []byte("0123456789abcdef0123456789abcdef")
  encryptor, err := encryption.NewAESEncryptor(key)
  if err != nil {
    t.Fatalf("NewAESEncryptor() error = %v", err)
  }

  _, err = encryptor.Decrypt("not-valid-base64!!!")
  if err == nil {
    t.Error("Decrypt() should return error for invalid base64")
  }

  _, err = encryptor.Decrypt("YQ==") // valid base64, but too short / invalid ciphertext
  if err == nil {
    t.Error("Decrypt() should return error for invalid ciphertext")
  }
}
