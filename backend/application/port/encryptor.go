package port

// TokenEncryptor defines operations for encrypting and decrypting tokens.
type TokenEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}
