package encrypt

import (
	"crypto/md5"
	"crypto/sha256"
	"github.com/pkg/errors"
	"github.com/spacemonkeygo/openssl"
)

type encrypt struct {
	*openssl.Cipher

	key, iv []byte
}

func New(name string, key, iv []byte) (*encrypt, error) {
	c, err := openssl.GetCipherByName(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting cipher failed")
	}

	return &encrypt{
		Cipher: c,
		key:    key,
		iv:     iv,
	}, nil
}

func (enc *encrypt) Encrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewEncryptionCipherCtx(enc.Cipher, nil, enc.key, enc.iv)
	if err != nil {
		return nil, err
	}

	result, err := ctx.EncryptUpdate(input)
	if err != nil {
		return nil, err
	}

	final, err := ctx.EncryptFinal()
	if err != nil {
		return nil, err
	}

	result = append(result, final...)
	return result, nil
}

func (enc *encrypt) Decrypt(input []byte) ([]byte, error) {
	ctx, err := openssl.NewDecryptionCipherCtx(enc.Cipher, nil, enc.key, enc.iv)
	if err != nil {
		return nil, err
	}

	result, err := ctx.DecryptUpdate(input)
	if err != nil {
		return nil, err
	}

	final, err := ctx.DecryptFinal()
	if err != nil {
		return nil, err
	}

	result = append(result, final...)
	return result, nil
}

func KeysFromString(str string) ([]byte, []byte) {
	b := []byte(str)
	key, iv := sha256.Sum256(b), md5.Sum(b)
	return key[:], iv[:]
}
