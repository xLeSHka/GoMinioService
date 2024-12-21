package crypto

import "testing"

func TestCrypto(t *testing.T) {
	text := "Hello, world!"
	key := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	var encryptedTest []byte

	t.Run("Encrypt", func(t *testing.T) {
		en, err := Encrypt([]byte(text), key)
		if err != nil {
			t.Fatal(err)
		}

		encryptedTest = en
	})

	t.Run("Decrypt", func(t *testing.T) {
		de, err := Decrypt(encryptedTest, key)
		if err != nil {
			t.Fatal(de)
		}

		if string(de) != text {
			t.Fatalf("Incorrect transcript. Received: %s, pending: %s", de, text)
		}
	})

	t.Run("Error encrypt", func(t *testing.T) {
		_, err := Encrypt([]byte{}, []byte{})
		if err == nil {
			t.Fatalf("Didn't get the expected error")
		}
	})

	t.Run("Error decrypt", func(t *testing.T) {
		_, err := Decrypt([]byte{}, []byte{})
		if err == nil {
			t.Fatalf("Didn't get the expected error")
		}
	})
}
