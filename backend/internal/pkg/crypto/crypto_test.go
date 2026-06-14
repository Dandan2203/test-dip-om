package crypto

import "testing"

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := "test-encryption-key"
	secret := "uXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX" // схоже на токен Monobank

	enc, err := Encrypt(secret, key)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == secret {
		t.Fatal("зашифрований текст збігається з вихідним")
	}

	dec, err := Decrypt(enc, key)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != secret {
		t.Fatalf("очікувано %q, отримано %q", secret, dec)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	enc, _ := Encrypt("secret", "key-one")
	if _, err := Decrypt(enc, "key-two"); err == nil {
		t.Fatal("очікувалася помилка для невірного ключа")
	}
}
