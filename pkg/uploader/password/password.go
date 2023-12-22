package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
)

var key = []byte("obfuscation-key-")

type Password string

//goland:noinspection GoMixedReceiverTypes
func (p Password) MarshalJSON() ([]byte, error) {
	if p.PlainText() == "" {
		return []byte(`""`), nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(p))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(p))

	return json.Marshal(ciphertext)
}

//goland:noinspection GoMixedReceiverTypes
func (p *Password) UnmarshalJSON(bs []byte) error {
	if string(bs) == `""` {
		*p = ""
		return nil
	}

	var ciphertext []byte
	if err := json.Unmarshal(bs, &ciphertext); err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	if len(ciphertext) < aes.BlockSize {
		return errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	*p = Password(ciphertext)
	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (p Password) String() string {
	return "[redacted]"
}

//goland:noinspection GoMixedReceiverTypes
func (p Password) PlainText() string {
	return string(p)
}
