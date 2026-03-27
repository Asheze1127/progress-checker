package service

import "testing"

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	hasher := &PasswordHasher{}

	hash, err := hasher.Hash("my-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if hash == "my-password" {
		t.Fatal("hash should not equal plain password")
	}

	err = hasher.Verify(hash, "my-password")
	if err != nil {
		t.Errorf("Verify() error = %v, want nil for correct password", err)
	}
}

func TestPasswordHasher_Verify_WrongPassword(t *testing.T) {
	hasher := &PasswordHasher{}

	hash, err := hasher.Hash("correct-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	err = hasher.Verify(hash, "wrong-password")
	if err == nil {
		t.Fatal("Verify() expected error for wrong password, got nil")
	}
}

func TestPasswordHasher_Hash_DifferentResultsForSameInput(t *testing.T) {
	hasher := &PasswordHasher{}

	hash1, _ := hasher.Hash("same-password")
	hash2, _ := hasher.Hash("same-password")

	if hash1 == hash2 {
		t.Error("bcrypt should generate different hashes for the same input (different salt)")
	}
}
